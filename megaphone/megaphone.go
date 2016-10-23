package megaphone

import (
	"fmt"
	"github.com/0x7fffffff/verbatim/model"
	"github.com/0x7fffffff/verbatim/persist"
	"log"
	"net"
)

func encId(enc model.Encoder) EncoderID {
	return EncoderID(enc.ID)
}

func networkId(enc model.Encoder) NetworkID {
	return NetworkID(enc.NetworkID)
}

func netId(n model.Network) NetworkID {
	return NetworkID(n.ID)
}

type MegaphoneListener interface {
	// Logged into encoder properly
	LoginSucceeded(enc model.Encoder)
	// Logged into encoder properly
	LoginFailed(enc model.Encoder)
	// Writing to an encoder failed for some reason
	UnexpectedDisconnect(enc model.Encoder)
	// An encoder was logged out
	Logout(enc model.Encoder)
}

func NotifyNetworkAdded(n model.Network) {
	networkAdded <- n
}

func NotifyNetworkRemoved(n model.Network) {
	networkRemoved <- n
}

func NotifyEncoderAdded(enc model.Encoder) {
	encoderAdded <- enc
}

func NotifyEncoderRemoved(enc model.Encoder) {
	encoderRemoved <- enc
}

func NotifyEncoderLogout(enc model.Encoder) {
	encoderRemoved <- enc
}

var l MegaphoneListener

func Start(ml MegaphoneListener) error {
	l = ml
	return setupEncoders()
}

// Crud notifications
var (
	networkAdded   = make(chan model.Network, 10)
	networkRemoved = make(chan model.Network, 10)
	encoderRemoved = make(chan model.Encoder, 10)
	encoderAdded   = make(chan model.Encoder, 10)
	encoderLogout  = make(chan model.Encoder, 10)
)

// Doing lookups on broadcasters
var (
	askBroadcaster   = make(chan NetworkID)
	giveBroadCasters = make(chan *NetworkBroadcaster)
)

type NetworkID int
type EncoderID int

// Send notifications (coming from Relay server)
var networkBroadcasters = make(map[NetworkID]*NetworkBroadcaster)

func GetBroadcasterForNetwork(id NetworkID) *NetworkBroadcaster {
	askBroadcaster <- id
	return <-giveBroadCasters
}

func setupEncoders() error {
	// networks, err := persist.GetNetworks()
	encoders, err := persist.GetEncoders()
	if err != nil {
		return err
	}
	networkBroadcasters = make(map[NetworkID]*NetworkBroadcaster)
	encoderFaulted := make(chan encoderIdPair)
	for _, encoder := range encoders {
		var broadcaster *NetworkBroadcaster
		if val, found := networkBroadcasters[networkId(encoder)]; found {
			broadcaster = val
		} else {
			broadcaster = makeBroadcaster(networkId(encoder), encoderFaulted)
			networkBroadcasters[networkId(encoder)] = broadcaster
			// Launch this off so that calls below don't block
			go broadcaster.serveConnection()
		}
		inbound := make(chan []byte)
		broadcaster.registerEncoderChan(encId(encoder), inbound)
		go handleEncoder(encoder, inbound, broadcaster)
	}
	daemonOfAwesome(networkBroadcasters, encoderFaulted)
	return fmt.Errorf("Close the daemon of awesome for some reason")
}

func daemonOfAwesome(broadcasters map[NetworkID]*NetworkBroadcaster, encoderFaulted chan encoderIdPair) {
	for {
		select {
		case id := <-askBroadcaster:
			giveBroadCasters <- broadcasters[id]
		case newNet := <-networkAdded:
			if _, found := broadcasters[netId(newNet)]; !found {
				b := makeBroadcaster(netId(newNet), encoderFaulted)
				broadcasters[netId(newNet)] = b
				go b.serveConnection()
			}
		case killNet := <-networkRemoved:
			if b, found := broadcasters[NetworkID(killNet.ID)]; found {
				b.destroy()
				delete(broadcasters, NetworkID(killNet.ID))
			}

		case enc := <-encoderRemoved:
			broadcasters[NetworkID(enc.NetworkID)].removeEncoder(EncoderID(enc.ID))

		case restartEnc := <-encoderFaulted:
			if b, found := broadcasters[restartEnc.network]; found {
				inbound := make(chan []byte)
				// If the encoder is already running, don't try to start it again
				if b.registerEncoderChan(restartEnc.encoder, inbound) == encoderDidExist {
					continue
				}
				// Refresh the info from the database
				enc, err := persist.GetEncoder(int(restartEnc.encoder))
				if err != nil {
					// Try to restart this at the next tick
					b.faultedEncoder <- EncoderID(enc.ID)
					continue
				}
				// Remove the encoder from the broadcaster if it dies
				go handleEncoder(*enc, inbound, b)
			}
		case newEnc, ok := <-encoderAdded:
			if !ok {
				log.Print("Closed network addition channel!")
				return
			}

			// Make sure we're adding to a network that exists
			if b, found := broadcasters[NetworkID(newEnc.NetworkID)]; !found {
				inbound := make(chan []byte)
				// If we are asked to add an existing encoder, then do nothing
				if b.registerEncoderChan(encId(newEnc), inbound) == encoderDidNotExist {
					// Remove the encoder from the broadcaster if it dies
					go handleEncoder(newEnc, inbound, b)
				} else {
					close(inbound)
				}
			}
		}
	}
}

const LINE_CUT_WIDTH = 32

func writeMessageSegmented(conn net.Conn, msg []byte) error {
	// Write message in chunks
	for i := 0; i*LINE_CUT_WIDTH < len(msg); i++ {
		begin := i * LINE_CUT_WIDTH
		end := (i + 1) * LINE_CUT_WIDTH
		if end > len(msg) {
			end = len(msg)
		}
		if _, err := conn.Write(msg[begin:end]); err != nil {
			return err
		}
	}
	return nil
}

func loginToEncoder(enc model.Encoder) (net.Conn, error) {
	conn, err := net.Dial("tcp", fmt.Sprint(enc.IPAddress, ":", enc.Port))
	if err != nil {
		return nil, err
	}
	if _, err = conn.Write([]byte(enc.Handle + "\n")); err != nil {
		return nil, err
	}
	if _, err = conn.Write([]byte(enc.Password + "\n")); err != nil {
		return nil, err
	}
	// TODO: Read response here?
	return conn, nil
}

func handleEncoder(enc model.Encoder, inbound chan []byte, n *NetworkBroadcaster) {
	conn, err := loginToEncoder(enc)
	if err != nil {
		// Login failed, remove it from the list of the things
		n.removeEncoder(encId(enc))
		// And then notify that login failed for the encoder
		// Allowing the user to try to relogin
		l.LoginFailed(enc)
		conn.Close()
		return
	}
	l.LoginSucceeded(enc)
	for {
		select {

		case msg, ok := <-inbound:
			if ok {
				err := writeMessageSegmented(conn, msg)
				if err != nil {
					// Close the connection
					conn.Close()
					// Signal to the broadcaster that we have an error
					// and will need to be restarted
					// Try to restart
					l.UnexpectedDisconnect(enc)
					n.faultedEncoder <- encId(enc)
					return
				}
			} else {
				conn.Close()
				l.Logout(enc)
				//
				return
			}
		}
	}
}

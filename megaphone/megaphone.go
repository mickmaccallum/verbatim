package megaphone

import (
	"fmt"
	"github.com/0x7fffffff/verbatim/model"
	"github.com/0x7fffffff/verbatim/persist"
	"log"
	"net"
)

func encId(enc model.Encoder) model.EncoderID {
	return model.EncoderID(enc.ID)
}

func networkId(enc model.Encoder) model.NetworkID {
	return model.NetworkID(enc.NetworkID)
}

func netId(n model.Network) model.NetworkID {
	return model.NetworkID(n.ID)
}

type MegaphoneListener interface {
	// Notify that the network has been added to the megaphone
	// NetworkAdded(net model.Network)
	// Notify that the network has been removed, along with all it's effects, from the megaphone
	// NetworkRemoved(net model.NetworkID)

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

func NotifyNetworkRemoved(n model.NetworkID) {
	networkRemoved <- n
}

func NotifyEncoderLogin(enc model.Encoder) {
	encoderAdded <- enc
}

func NotifyEncoderLogout(enc model.Encoder) {
	encoderRemoved <- enc
}

var relay MegaphoneListener

func Start(ml MegaphoneListener) error {
	relay = ml
	return setupEncoders()
}

// Crud notifications
var (
	networkAdded   = make(chan model.Network, 10)
	networkRemoved = make(chan model.NetworkID, 10)
	encoderRemoved = make(chan model.Encoder, 10)
	encoderAdded   = make(chan model.Encoder, 10)
	encoderLogout  = make(chan model.Encoder, 10)
)

// Doing lookups on broadcasters
var (
	askBroadcaster       = make(chan model.NetworkID)
	giveBroadCasters     = make(chan *NetworkBroadcaster)
	getConnectedEncoders = make(chan model.NetworkID)
	connectedEncoders    = make(chan []model.EncoderID)
)

// Send notifications (coming from Relay server)
var networkBroadcasters = make(map[model.NetworkID]*NetworkBroadcaster)

func GetBroadcasterForNetwork(id model.NetworkID) *NetworkBroadcaster {
	askBroadcaster <- id
	return <-giveBroadCasters
}

func GetConnectedEncoders(id model.NetworkID) []model.EncoderID {
	getConnectedEncoders <- id
	return <-connectedEncoders
}

func setupEncoders() error {
	networks, err := persist.GetNetworks()
	if err != nil {
		return nil
	}

	encoders, err := persist.GetEncoders()
	if err != nil {
		return err
	}

	encoderFaulted := make(chan encoderIdPair)
	networkBroadcasters = make(map[model.NetworkID]*NetworkBroadcaster)
	for _, n := range networks {
		broadcaster := makeBroadcaster(n.ID, encoderFaulted)
		networkBroadcasters[n.ID] = broadcaster
		// Launch this off so that calls below don't block
		go broadcaster.serveConnection()
	}
	for _, encoder := range encoders {
		broadcaster := networkBroadcasters[networkId(encoder)]
		inbound := make(chan []byte)
		broadcaster.registerEncoderChan(encId(encoder), inbound)
		go handleEncoder(encoder, inbound, broadcaster)
	}
	daemonOfAwesome(networkBroadcasters, encoderFaulted)
	return fmt.Errorf("Closed the daemon of awesome for some reason")
}

func daemonOfAwesome(broadcasters map[model.NetworkID]*NetworkBroadcaster, encoderFaulted chan encoderIdPair) {
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
			if b, found := broadcasters[killNet]; found {
				b.destroy()
				delete(broadcasters, killNet)
			}
		case netId := <-getConnectedEncoders:
			if b, found := broadcasters[netId]; found {
				connectedEncoders <- b.getConnectedEncoderIds()
			} else {
				connectedEncoders <- make([]model.EncoderID, 0)
			}
		case enc := <-encoderRemoved:
			broadcasters[model.NetworkID(enc.NetworkID)].removeEncoder(model.EncoderID(enc.ID))

		// TODO: This code isn't currently being used, but could be used if we need something like this in the future
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
					b.faultedEncoder <- model.EncoderID(enc.ID)
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
			if b, found := broadcasters[model.NetworkID(newEnc.NetworkID)]; found {
				inbound := make(chan []byte)
				// If we are asked to add an existing encoder, then do nothing
				if b.registerEncoderChan(encId(newEnc), inbound) == encoderDidNotExist {
					// Remove the encoder from the broadcaster if it dies
					log.Println("Started new encoder")
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
		relay.LoginFailed(enc)
		// conn.Close()
		return
	}
	relay.LoginSucceeded(enc)
	for {
		select {

		case msg, ok := <-inbound:
			if ok {
				err := writeMessageSegmented(conn, msg)
				if err != nil {
					// Signal to the broadcaster that we have an error
					relay.UnexpectedDisconnect(enc)
					n.removeEncoder(enc.ID)
					n.faultedEncoder <- encId(enc)
					return
				}
			} else {
				log.Println("Encoder removed")
				conn.Close()
				relay.Logout(enc)
				return
			}
		}
	}
}

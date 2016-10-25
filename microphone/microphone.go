package microphone

import (
	"fmt"
	"log"
	"math"
	"net"
	"time"

	"github.com/0x7fffffff/verbatim/megaphone"
	"github.com/0x7fffffff/verbatim/model"
	"github.com/0x7fffffff/verbatim/persist"
)

type CaptionerID struct {
	IPAddr    string
	NumConn   int
	NetworkID NetworkID
}

type NetworkID int

func (c CaptionerID) String() string {
	return fmt.Sprint(c.IPAddr, ":", c.NumConn)
}

// The connection information for a given captioner.
type CaptionerInfo struct {
	CaptionerID
	network model.Network
}

type CaptionerStatus struct {
	CaptionerInfo
	lastActivity time.Time
}

// These are the events that the server using this will be notified of
type CaptionListener interface {

	// When a captioner connects
	Connected(ci CaptionerInfo)

	// Report that a captioner has been disconnected
	Disconnected(ci CaptionerInfo)

	// Report that a given network was unable to listen to a given port
	NetworkListenFailed(network model.Network)

	// Report that the server for the network was able to listen
	NetworkListenSucceeded(network model.Network)

	// Report that a captioner has been successfully muted
	Muted(ci CaptionerInfo)

	// Report that a captioner has been successfully unmuted
	Unmuted(ci CaptionerInfo)

	// Ask the relay server for a network writer
	GetBroadcaster(network model.Network) megaphone.NetworkBroadcaster
}

// This is our reference to the delegate methods from the relay server
var relay CaptionListener

// Listen for TCP connections
func Start(listener CaptionListener) error {
	// Get this ready first
	go maintainListenerState()
	relay = listener
	networks, err := persist.GetNetworks()
	if err != nil {
		return err
	}
	for _, n := range networks {
		addNetwork <- n
	}
	return nil
}

// TODO: Error handling here?
func AddNetwork(n model.Network) {
	// Hrm...
	// go listenForNetwork(n)
}

func MuteCaptioner(n model.Network, captionerId int) {
}

// Paired channels
var (
	addNetwork          = make(chan model.Network, 10)
	networkWasAdded     = make(chan NetworkListener, 10)
	networkListenFailed = make(chan NetworkID, 10)
	rmNetwork           = make(chan model.Network, 10)
	addCaptioner        = make(chan MuteCell, 10)
	rmCaptioner         = make(chan MuteCell, 10)
	muteCaptioner       = make(chan CaptionerID, 10)
	unmuteCaptioner     = make(chan CaptionerID, 10)
)

// Listeners by network
type NetworkListener struct {
	id       NetworkID
	listener *net.Listener
}

var askStatus = make(chan struct{})
var getStatus = make(chan map[model.Network]map[CaptionerInfo]struct{})

// This function is the sole arbiter of state for these stats
func maintainListenerState() {
	// Networks
	networks := make(map[NetworkID]NetworkListener)
	// Caption listeners

	// TODO: Uncomment this, and use it in the code below
	// writers := make(map[CaptionerID]MuteCell)

	for {
		select {
		case n := <-addNetwork:
			if ln, err := attemptListen(n); err != nil {
				go listenForNetwork(n, ln)
				networks[NetworkID(n.ID)] = NetworkListener{NetworkID(n.ID), &ln}
				relay.NetworkListenSucceeded(n)
			} else {
				relay.NetworkListenFailed(n)
			}
			/*
				case l := <-addCaptioner:
					network := networks[l.]
					network[l] = struct{}{}
					relay.Connected(l, l.network)

				case n := <-rmNetwork:
					if nln, found := networks[n.ID]; found {
						nln.listener.Close()
					} else {

					}
					delete(networks, n)

				case l := <-rmCaptioner:
					delete(networks[l.network], l)
					relay.Disconnected(l, l.network)

				case m := <-muteCaptioner:
			*/

		}
	}
}

func attemptListen(n model.Network) (net.Listener, error) {
	return net.Listen("tcp", fmt.Sprint(":", n.ListeningPort))
}

// TODO A way to report failed listers?
func listenForNetwork(n model.Network, ln net.Listener) {
	var connsPerIP = make(map[string]int)
	// Add the network to the list of networks in use.
	addNetwork <- n
	broadcaster := relay.GetBroadcaster(n)
	for {
		conn, er := ln.Accept()
		err := er.(net.Error)
		if err != nil {
			if err.Temporary() {
				log.Println(err)
				continue
			} else {
				// TODO: Signal this listener has been closed
				return
			}
		}

		// Make sure that a given captioner has a way of being identified
		key := conn.RemoteAddr().String()
		if val, found := connsPerIP[key]; found {
			connsPerIP[key] = (val + 1) % math.MaxInt32
		} else {
			connsPerIP[key] = 1
		}
		val := connsPerIP[key]
		writer := makeMuteCell(&broadcaster, CaptionerID{
			NumConn:   val,
			IPAddr:    conn.RemoteAddr().String(),
			NetworkID: NetworkID(n.ID),
		})
		go handleCaptioner(conn, writer)
	}
}

// Basic demoable state.
func handleCaptioner(c net.Conn, writer *MuteCell) {
	// Notify that we have a new listener
	// addListenerChan <-
	// Keep a buffer of 1KiB per captioner
	buf := make([]byte, 1024)
	for {
		c.SetReadDeadline(time.Now().Add(time.Minute * 10))
		n, err := c.Read(buf)
		if err != nil || n == 0 {
			c.Close()
			rmCaptioner <- *writer
			log.Println(err.Error())
			break
		}
		// Copy the data to make sure the
		// byte slice doesn't get changed under the encoder sending
		// it out.
		message := make([]byte, n)
		copy(message, buf[0:n])
		writer.Write(message)
		// Send any recieved bytes to the relay server
	}
	log.Printf("Connection from %v closed.", c.RemoteAddr())
}

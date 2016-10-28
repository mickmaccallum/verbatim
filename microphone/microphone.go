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

type CaptionerStatus struct {
	CaptionerID
	lastActivity time.Time
}

// These are the events that the server using this will be notified of
type RelayListener interface {

	// When a captioner connects
	Connected(ci CaptionerID)

	// Report that a captioner has been disconnected
	Disconnected(ci CaptionerID)

	// Report that a given network was unable to listen to a given port
	NetworkListenFailed(network model.Network)

	// Report that the server for the network was able to listen
	NetworkListenSucceeded(network model.Network)

	// Report that a captioner has been successfully muted
	Muted(ci CaptionerID)

	// Report that a captioner has been successfully unmuted
	Unmuted(ci CaptionerID)

	// Ask the relay server for a network writer
	GetBroadcaster(network model.Network) megaphone.NetworkBroadcaster
}

// This is our reference to the delegate methods from the relay server
var relay RelayListener

// Listen for TCP connections
func Start(listener RelayListener) error {
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

// Listen on this network's port, and track listeners over time
func AddNetwork(n model.Network) {
	addNetwork <- n
}

// Stop listening for captioners on this network, frees up the port assigned to the network
func RemoveNetwork(id NetworkID) {
	rmNetwork <- id
}

// Forcibly disconnect this captioner, so they cannot send captions.
func RemoveCaptioner(id CaptionerID) {
	rmCaptioner <- id
}

// Mute the captioner with the associated id.
func MuteCaptioner(id CaptionerID) {
	muteCaptioner <- id
}

// Unmute the captioner with the associated id.
func UnmuteCaptioner(id CaptionerID) {
	unmuteCaptioner <- id
}

// Paired channels
var (
	addNetwork      = make(chan model.Network, 10)
	rmNetwork       = make(chan NetworkID, 10)
	captionerAdded  = make(chan CaptionListener, 10) // Notify a new captioner has been added
	rmCaptioner     = make(chan CaptionerID, 10)
	muteCaptioner   = make(chan CaptionerID, 10)
	unmuteCaptioner = make(chan CaptionerID, 10)
)

// Listeners by network
type NetworkListener struct {
	id       NetworkID
	listener net.Listener
}

type CaptionListener struct {
	NetId NetworkID
	conn  net.Conn
	cell  *MuteCell
}

// This function is the sole arbiter of state for these stats
func maintainListenerState() {
	// Networks
	networks := make(map[NetworkID]NetworkListener)
	// Caption listeners
	listeners := make(map[CaptionerID]CaptionListener)
	listenersByNetwork := make(map[NetworkID][]CaptionListener)

	// TODO: Uncomment this, and use it in the code below
	// writers := make(map[CaptionerID]MuteCell)

	for {
		select {
		case n := <-addNetwork:
			if ln, err := attemptListen(n); err != nil {
				networks[NetworkID(n.ID)] = NetworkListener{NetworkID(n.ID), ln}
				go listenForNetwork(n, ln)
				relay.NetworkListenSucceeded(n)
			} else {
				relay.NetworkListenFailed(n)
			}
		case cl := <-captionerAdded:
			// Note: This is only fired when the network listener wants to let us know we have a new captioner
			listeners[CaptionerID(cl.cell.id)] = cl
			if arr, found := listenersByNetwork[NetworkID(cl.NetId)]; found {
				arr = append(arr, cl)
			} else {
				arr = []CaptionListener{cl}
			}
			relay.Connected(cl.cell.id)
		case rmId := <-rmCaptioner:
			if cl, found := listeners[rmId]; found {
				cl.cell.Mute()
				cl.conn.Close()
				relay.Disconnected(rmId)
			}
		case rmId := <-rmNetwork:
			if network, found := networks[NetworkID(rmId)]; found {
				// Tear down all the caption side stuff when a network is to be removed
				// Keep from getting new connections
				network.listener.Close()
				for _, captioner := range listenersByNetwork[rmId] {
					delete(listeners, captioner.cell.id)
					// Mute the cell
					captioner.cell.Mute()
					// Close the connection
					captioner.conn.Close()
				}
				delete(listenersByNetwork, rmId)
			}
		case muteId := <-muteCaptioner:
			if cl, found := listeners[muteId]; found {
				cl.cell.Mute()
				relay.Muted(muteId)
			}
		case unmuteId := <-unmuteCaptioner:
			if cl, found := listeners[unmuteId]; found {
				cl.cell.Unmute()
				relay.Unmuted(unmuteId)
			}
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
		captionerAdded <- CaptionListener{
			conn: conn,
			cell: writer,
		}
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
			rmCaptioner <- writer.id
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
	// log.Printf("Connection from %v closed.", c.RemoteAddr())
}

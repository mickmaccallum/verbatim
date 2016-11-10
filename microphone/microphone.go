package microphone

import (
	"fmt"
	"github.com/0x7fffffff/verbatim/states"
	"log"
	"math"
	"net"
	"time"

	"github.com/0x7fffffff/verbatim/megaphone"
	"github.com/0x7fffffff/verbatim/model"
	"github.com/0x7fffffff/verbatim/persist"
)

// The connection information for a given captioner.
type CaptionerStatus struct {
	model.CaptionerID
	states.Captioner
}

// These are the events that the server using this will be notified of
type RelayListener interface {

	// Report that a given network was unable to listen to a given port
	NetworkListenFailed(network model.Network)

	// Report that the server for the network was able to listen
	NetworkListenSucceeded(network model.Network)

	// Network disconnected
	NetworkRemoved(network model.NetworkID)

	// When a captioner connects
	Connected(ci model.CaptionerID)

	// Report that a captioner has been disconnected
	Disconnected(ci model.CaptionerID)

	// Report that a captioner has been successfully muted
	Muted(ci model.CaptionerID)

	// Report that a captioner has been successfully unmuted
	Unmuted(ci model.CaptionerID)

	// Ask the relay server for a network writer
	GetBroadcaster(network model.Network) *megaphone.NetworkBroadcaster
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
func RemoveNetwork(id model.NetworkID) {
	rmNetwork <- id
}

// Forcibly disconnect this captioner, so they cannot send captions.
func RemoveCaptioner(id model.CaptionerID) {
	rmCaptioner <- id
}

// Mute the captioner with the associated id.
func MuteCaptioner(id model.CaptionerID) {
	muteCaptioner <- id
}

// Unmute the captioner with the associated id.
func UnmuteCaptioner(id model.CaptionerID) {
	unmuteCaptioner <- id
}

func GetConnectedCaptioners(m model.Network) []CaptionerStatus {
	askCaptioners <- m.ID
	return <-captionerStats
}

// Paired channels
var (
	addNetwork      = make(chan model.Network, 10)
	rmNetwork       = make(chan model.NetworkID, 10)
	captionerAdded  = make(chan CaptionListener, 10) // Notify a new captioner has been added
	rmCaptioner     = make(chan model.CaptionerID, 10)
	muteCaptioner   = make(chan model.CaptionerID, 10)
	unmuteCaptioner = make(chan model.CaptionerID, 10)
	askCaptioners   = make(chan model.NetworkID, 10)
	captionerStats  = make(chan []CaptionerStatus, 10)
)

// Listeners by network
type NetworkListener struct {
	id       model.NetworkID
	listener net.Listener
}

type CaptionListener struct {
	NetId model.NetworkID
	conn  net.Conn
	cell  *MuteCell
}

// This function is the sole arbiter of state for these stats
func maintainListenerState() {
	// Networks
	networks := make(map[model.NetworkID]NetworkListener)
	// Caption listeners
	listeners := make(map[model.CaptionerID]CaptionListener)
	listenersByNetwork := make(map[model.NetworkID][]CaptionListener)

	for {
		select {
		case n := <-addNetwork:
			if ln, err := attemptListen(n); err == nil {
				networks[model.NetworkID(n.ID)] = NetworkListener{model.NetworkID(n.ID), ln}
				go listenForNetwork(n, ln)
				relay.NetworkListenSucceeded(n)
			} else {
				relay.NetworkListenFailed(n)
			}
		case cl := <-captionerAdded:
			// Note: This is only fired when the network listener wants to let us know we have a new captioner
			listeners[model.CaptionerID(cl.cell.id)] = cl
			if arr, found := listenersByNetwork[model.NetworkID(cl.NetId)]; found {
				arr = append(arr, cl)
				if len(arr) == 1 {
					cl.cell.Unmute()
					relay.Unmuted(cl.cell.id)
				}
			} else {
				arr = []CaptionListener{cl}
				cl.cell.Unmute()
			}
			relay.Connected(cl.cell.id)
		case rmId := <-rmCaptioner:
			if cl, found := listeners[rmId]; found {
				cl.cell.Mute()
				cl.conn.Close()
				relay.Disconnected(rmId)
				if arr, found := listenersByNetwork[cl.NetId]; found && len(arr) == 1 {
					// Make sure the remaining captioner is unmuted
					arr[0].cell.Unmute()
				}
			}
		case rmId := <-rmNetwork:
			if network, found := networks[rmId]; found {
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
				relay.NetworkRemoved(rmId)
			}
		case netId := <-askCaptioners:
			if cells, found := listenersByNetwork[netId]; found {
				stats := make([]CaptionerStatus, 0)
				for _, cl := range cells {
					cl.cell.cellMux.Lock()
					if cl.cell.isMute {
						stats = append(stats, CaptionerStatus{
							cl.cell.id,
							states.CaptionerMuted,
						})
					} else {
						stats = append(stats, CaptionerStatus{
							cl.cell.id,
							states.CaptionerUnmuted,
						})
					}
					cl.cell.cellMux.Unlock()
				}
				captionerStats <- stats
			} else {
				captionerStats <- nil
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
		if er != nil {
			log.Println("Connection failed:", er.Error())
			if err := er.(net.Error); err != nil {
				if err.Temporary() {
					log.Println(err)
					continue
				} else {
					// TODO: Signal error here.
					return
				}
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
		writer := makeMuteCell(broadcaster, model.CaptionerID{
			NumConn:   val,
			IPAddr:    conn.RemoteAddr().String(),
			NetworkID: model.NetworkID(n.ID),
		})
		captionerAdded <- CaptionListener{
			conn: conn,
			cell: writer,
		}
		go handleCaptioner(conn, writer)
	}
}

func handleCaptioner(c net.Conn, writer *MuteCell) {
	// Notify that we have a new listener
	// addListenerChan <-
	// Keep a buffer of 1KiB per captioner
	buf := make([]byte, 1024)
	log.Println("Am listening to captioner")
	for {
		c.SetReadDeadline(time.Now().Add(time.Minute * 10))
		n, err := c.Read(buf)
		if err != nil || n == 0 {
			log.Println("Disconnected from Captioner")
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

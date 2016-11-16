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

var addNetwork = make(chan model.Network, 10)

// Listen on this network's port, and track listeners over time
func AddNetwork(n model.Network) {
	addNetwork <- n
}

var rmNetwork = make(chan model.NetworkID, 10)

// Stop listening for captioners on this network, frees up the port assigned to the network
func RemoveNetwork(id model.NetworkID) {
	rmNetwork <- id
}

var (
	askNetworks = make(chan struct{})
	gotNetworks = make(chan map[model.NetworkID]bool)
)

var (
	askPortChange = make(chan struct {
		model.NetworkID
		int
	})
	couldStartPortChange = make(chan error)
)

func AttemptPortChange(id model.NetworkID, int newPort) error {
	askPortChange <- struct {
		model.NetworkID
		int
	}{id, newPort}
	return <-couldStartPortChange
}

// Returns all the successfully connected networks
func GetListeningNetworks() map[model.NetworkID]bool {
	askNetworks <- struct{}{}
	return <-gotNetworks
}

var rmCaptioner = make(chan model.CaptionerID, 10)

// Forcibly disconnect this captioner, so they cannot send captions.
func RemoveCaptioner(id model.CaptionerID) {
	rmCaptioner <- id
}

var muteCaptioner = make(chan model.CaptionerID, 10)

// Mute the captioner with the associated id.
func MuteCaptioner(id model.CaptionerID) {
	muteCaptioner <- id
}

var unmuteCaptioner = make(chan model.CaptionerID, 10)

// Unmute the captioner with the associated id.
func UnmuteCaptioner(id model.CaptionerID) {
	unmuteCaptioner <- id
}

var (
	askCaptioners  = make(chan model.NetworkID, 10)
	captionerStats = make(chan []CaptionerStatus, 10)
)

func GetConnectedCaptioners(m model.Network) []CaptionerStatus {
	askCaptioners <- m.ID
	return <-captionerStats
}

// Captioner channels
var (
	tryAddCaptioner   = make(chan CaptionListener) // Notify a new captioner has been added
	couldAddCaptioner = make(chan error)
	errNetworkClosed  = fmt.Errorf("The network's port was closed")
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
				networks[n.ID] = NetworkListener{model.NetworkID(n.ID), ln}
				listenersByNetwork[n.ID] = make([]CaptionListener, 0)
				go listenForNetwork(n, ln)
				relay.NetworkListenSucceeded(n)
			} else {
				relay.NetworkListenFailed(n)
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
		case <-askNetworks:
			connectedNetworks := make(map[model.NetworkID]bool)
			for _, n := range networks {
				connectedNetworks[n.id] = true
			}
			gotNetworks <- connectedNetworks
		case cl := <-tryAddCaptioner:
			// Note: This is only fired when the network listener wants to let us know we have a new captioner
			listeners[model.CaptionerID(cl.cell.id)] = cl
			if arr, found := listenersByNetwork[model.NetworkID(cl.NetId)]; found {
				arr = append(arr, cl)
				if len(arr) == 1 {
					cl.cell.Unmute()
				}
				listenersByNetwork[cl.NetId] = arr
				relay.Connected(cl.cell.id)
				cl.cell.cellMux.Lock()
				if cl.cell.isMute {
					relay.Muted(cl.cell.id)
				} else {
					relay.Unmuted(cl.cell.id)
				}
				cl.cell.cellMux.Unlock()
				couldAddCaptioner <- nil
			} else {
				couldAddCaptioner <- fmt.Errorf("")
			}
		case rmId := <-rmCaptioner:
			if cl, found := listeners[rmId]; found {
				cl.cell.Mute()
				cl.conn.Close()
				relay.Disconnected(rmId)
				if arr, found := listenersByNetwork[cl.NetId]; found && len(arr) == 1 {
					// Make sure the remaining captioner is unmuted
					arr[0].cell.Unmute()
				}
				// Remove the listener from the list of listeners
				delete(listeners, rmId)
				toSplice := listenersByNetwork[rmId.NetworkID]
				if len(toSplice) > 1 {
					for i, captioner := range toSplice {
						if captioner.cell.id == rmId {
							// Using a 0-valued item to make sure that storage doesn't hold onto the conn
							toSplice[i] = CaptionListener{}
							listenersByNetwork[rmId.NetworkID] = append(toSplice[:i], toSplice[i+1:]...)
							break
						}
					}
				} else {
					delete(listenersByNetwork, rmId.NetworkID)
				}
			}
		case netId := <-askCaptioners:
			log.Println("Check captioners for network:", netId)
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

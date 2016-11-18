package microphone

import (
	"fmt"
	"github.com/0x7fffffff/verbatim/states"
	"net"

	"github.com/0x7fffffff/verbatim/megaphone"
	"github.com/0x7fffffff/verbatim/model"
	"github.com/0x7fffffff/verbatim/persist"
)

// The connection information for a given captioner.
type CaptionerStatus struct {
	id    model.CaptionerID
	state states.Captioner
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
		port int
	})
	couldStartPortChange = make(chan error)
)

func AttemptPortChange(id model.NetworkID, newPort int) error {
	askPortChange <- struct {
		model.NetworkID
		port int
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

// This function is the sole arbiter of state for these stats
func maintainListenerState() {
	// Networks
	networks := make(map[model.NetworkID]networkListeningServer)
	// Caption listeners
	// listeners := make(map[model.CaptionerID]CaptionListener)
	// listenersByNetwork := make(map[model.NetworkID][]CaptionListener)

	for {
		select {
		case n := <-addNetwork:
			if srv, err := tryMakeNetworkListener(n); err != nil {
				relay.NetworkListenFailed(n)
			} else {
				go srv.serve()
				relay.NetworkListenSucceeded(n)
			}
		case rmId := <-rmNetwork:
			if network, found := networks[rmId]; found {
				// Tear down all the caption side stuff when a network is to be removed
				// Keep from getting new connections
				// network.listener.Close()
				network.Close()
				delete(networks, rmId)
			}
		case <-askNetworks:
			connectedNetworks := make(map[model.NetworkID]bool)
			for _, n := range networks {
				connectedNetworks[n.network.ID] = true
			}
			gotNetworks <- connectedNetworks
		case info := <-askPortChange:
			couldStartPortChange <- networks[info.NetworkID].TryChangePort(info.port)
		case rmId := <-rmCaptioner:
			networks[rmId.NetworkID].RemoveCaptioner(rmId)
		case netId := <-askCaptioners:
			captionerStats <- networks[netId].GetConnectedCaptioners()
		case muteId := <-muteCaptioner:
			if n, found := networks[muteId.NetworkID]; found {
				n.MuteCaptioner(muteId)
			}
		case unmuteId := <-unmuteCaptioner:
			if n, found := networks[unmuteId.NetworkID]; found {
				n.UnmuteCaptioner(unmuteId)
			}
		}
	}
}

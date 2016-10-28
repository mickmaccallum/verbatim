package relay

import (
	"log"

	"github.com/0x7fffffff/verbatim/dashboard"
	"github.com/0x7fffffff/verbatim/megaphone"
	"github.com/0x7fffffff/verbatim/microphone"
	"github.com/0x7fffffff/verbatim/model"
	"github.com/0x7fffffff/verbatim/states"
)

// TODO: maintain a source of truth state that can be handed to web pages

type encoderListener struct{}

func (e encoderListener) NetworkListenFailed(network model.Network) {
	dashboard.NetworkPortStateChanged(network, states.NetworkListenFailed)
}

func (e encoderListener) NetworkListenSucceeded(network model.Network) {
	dashboard.NetworkPortStateChanged(network, states.NetworkListening)
}

func (e encoderListener) NetworkRemoved(network model.NetworkID) {
	// FIXME: Need to figure out a better way for this.
	var netModel = model.Network{
		ID: network,
	}
	dashboard.NetworkPortStateChanged(netModel, states.NetworkClosed)
}

func (e encoderListener) Connected(ci model.CaptionerID) {
	dashboard.CaptionerStateChanged(ci, states.CaptionerConnected)
}

func (e encoderListener) Disconnected(ci model.CaptionerID) {
	dashboard.CaptionerStateChanged(ci, states.CaptionerMuted)
}

func (e encoderListener) Muted(ci model.CaptionerID) {
	dashboard.CaptionerStateChanged(ci, states.CaptionerMuted)
}

func (e encoderListener) Unmuted(ci model.CaptionerID) {
	dashboard.CaptionerStateChanged(ci, states.CaptionerUnmuted)
}

func (e encoderListener) GetBroadcaster(network model.Network) *megaphone.NetworkBroadcaster {
	return megaphone.GetBroadcasterForNetwork(network.ID)
}

// ActivityType The type of a given activity
type ActivityType int

// NotifyPortAdded lint
func NotifyPortAdded(portNum int, n model.Network) {
	log.Println("Port added", portNum, n.Name)
}

// Start I'm a stub.
func Start() {
	go dashboard.Start(nil)
	go microphone.Start(encoderListener{})
	go megaphone.Start(nil)

}

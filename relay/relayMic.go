package relay

import (
	"github.com/0x7fffffff/verbatim/dashboard"
	"github.com/0x7fffffff/verbatim/megaphone"
	"github.com/0x7fffffff/verbatim/model"
	"github.com/0x7fffffff/verbatim/states"
)

// TODO: maintain a source of truth state that can be handed to web pages
type micListener struct{}

func (e micListener) NetworkListenFailed(network model.Network) {
	dashboard.NetworkPortStateChanged(network, states.NetworkListenFailed)
}

func (e micListener) NetworkListenSucceeded(network model.Network) {
	dashboard.NetworkPortStateChanged(network, states.NetworkListening)
}

func (e micListener) NetworkRemoved(network model.NetworkID) {
	// FIXME: Need to figure out a better way for this.
	var netModel = model.Network{
		ID: network,
	}
	dashboard.NetworkPortStateChanged(netModel, states.NetworkClosed)
}

func (e micListener) Connected(ci model.CaptionerID) {
	dashboard.CaptionerStateChanged(ci, states.CaptionerConnected)
}

func (e micListener) Disconnected(ci model.CaptionerID) {
	dashboard.CaptionerStateChanged(ci, states.CaptionerMuted)
}

func (e micListener) Muted(ci model.CaptionerID) {
	dashboard.CaptionerStateChanged(ci, states.CaptionerMuted)
}

func (e micListener) Unmuted(ci model.CaptionerID) {
	dashboard.CaptionerStateChanged(ci, states.CaptionerUnmuted)
}

func (e micListener) GetBroadcaster(network model.Network) *megaphone.NetworkBroadcaster {
	return megaphone.GetBroadcasterForNetwork(network.ID)
}

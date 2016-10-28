package relay

import (
	"github.com/0x7fffffff/verbatim/megaphone"
	"github.com/0x7fffffff/verbatim/microphone"
	"github.com/0x7fffffff/verbatim/model"
)

// TODO: Track nasty ball of state
// TODO: Decide whether dashboard or relay will talk to database

type dashboardListener struct{}

// Add network to database and relay-based servers
func (dl dashboardListener) AddNetwork(n model.Network) {
	// Add the megaphone network first,
	// so that we know that will be able to get a broadcaster
	// when we add the network to the microphone package
	megaphone.NotifyNetworkAdded(n)
	microphone.AddNetwork(n)
}

// Remove a network and *all* of it's encoders from the
// database and traffic
func (dl dashboardListener) RemoveNetwork(id model.NetworkID) {
	// Remove the network first from listening, and then from sending
	microphone.RemoveNetwork(id)
	megaphone.NotifyNetworkRemoved(id)
}

// Add encoder to it's network
func (dl dashboardListener) AddEncoder(enc model.Encoder) {
	megaphone.NotifyEncoderAdded(enc)
}

// Logout encoder
func (dl dashboardListener) LogoutEncoder(enc model.Encoder) {
	megaphone.NotifyEncoderLogout(enc)
}

// Remove encoder from database and from encoder
func (dl dashboardListener) DeleteEncoder(enc model.Encoder) {
	megaphone.NotifyEncoderRemoved(enc)
}

// Mute a captioner to keep them from being able to
// send data to the encoders
func (dl dashboardListener) MuteCaptioner(id model.CaptionerID) {
	microphone.MuteCaptioner(id)
}

// Unmute a captioner, allowing them to send data to encoders
func (dl dashboardListener) UnmuteCaptioner(id model.CaptionerID) {
	microphone.UnmuteCaptioner(id)
}

// Remove a captioner, forcibly disconnecting them
func (dl dashboardListener) RemoveCaptioner(id model.CaptionerID) {
	microphone.RemoveCaptioner(id)
}

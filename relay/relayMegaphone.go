package relay

import (
	"github.com/0x7fffffff/verbatim/dashboard"
	"github.com/0x7fffffff/verbatim/model"
	"github.com/0x7fffffff/verbatim/states"
)

type encoderListener struct{}

func (e encoderListener) LoggingIn(enc model.Encoder) {
	dashboard.EncoderStateChanged(enc, states.EncoderConnecting)
}
// Logged into encoder properly
func (e encoderListener) LoginSucceeded(enc model.Encoder) {
	dashboard.EncoderStateChanged(enc, states.EncoderConnected)
}

// Logging into an encoder failed
func (e encoderListener) LoginFailed(enc model.Encoder) {
	dashboard.EncoderStateChanged(enc, states.EncoderAuthFailure)
}

// Writing to an encoder failed for some reason
func (e encoderListener) UnexpectedDisconnect(enc model.Encoder) {
	dashboard.EncoderStateChanged(enc, states.EncoderFaulted)
}

// An encoder was logged out
func (e encoderListener) Logout(enc model.Encoder) {
	dashboard.EncoderStateChanged(enc, states.EncoderDisconnected)
}

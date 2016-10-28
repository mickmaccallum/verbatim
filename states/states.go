package states

// Represents the meaningful states of an Encoder
type Encoder int

const (
	// EncoderConnected Connected
	EncoderConnected Encoder = iota

	// EncoderConnecting not connected yet...
	EncoderConnecting

	// EncoderAuthFailure wrong credentials...
	EncoderAuthFailure

	// EncoderFaulted write failures happening, backing off.
	EncoderFaulted

	// EncoderDisconnected Disconnected (default state)
	EncoderDisconnected
)

// Captioner represents the states a captioner can be in.
type Captioner int

const (
	// A captioner has connected to our network
	CaptionerConnected Captioner = iota
	// A captioner has been disconnected from a network listener
	CaptionerDisconnected
	// A captioner has been muted
	CaptionerMuted
	// A captioner has been unmuted
	CaptionerUnmuted
)

// Network represents the state that the network
// listening on a port can be in
type Network int

const (
	// Network is in the process of connecting to a port
	Connecting Network = iota

	// Network has successfully connected to a port
	Listening

	// Network failed to listen on a port
	Failed

	// Network was told to stop listening
	Closed

	// Network was deleted
	Deleted
)

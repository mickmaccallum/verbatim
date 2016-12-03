package states

// Encoder Represents the meaningful states of an Encoder
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
	// CaptionerConnected A captioner has connected to our network
	CaptionerConnected Captioner = iota
	// CaptionerDisconnected A captioner has been disconnected from a network listener
	CaptionerDisconnected
	// CaptionerMuted A captioner has been muted
	CaptionerMuted
	// CaptionerUnmuted A captioner has been unmuted
	CaptionerUnmuted
)

// Network represents the state that the network
// listening on a port can be in
type Network int

const (
	// NetworkConnecting Network is in the process of connecting to a port
	NetworkConnecting Network = iota

	// NetworkListening Network has successfully connected to a port
	NetworkListening

	// NetworkListenFailed Network failed to listen on a port
	NetworkListenFailed

	// NetworkClosed Network was told to stop listening
	NetworkClosed

	// NetworkDeleted Network was deleted
	NetworkDeleted
)

package dashboard

import (
	"net/http"

	"github.com/0x7fffffff/verbatim/microphone"
	"github.com/0x7fffffff/verbatim/model"
	"github.com/0x7fffffff/verbatim/persist"
	"github.com/0x7fffffff/verbatim/states"
	"github.com/gorilla/sessions"
	"github.com/michaeljs1990/sqlitestore"
)

var store *sqlitestore.SqliteStore

func init() {
	var err error
	store, err = sqlitestore.NewSqliteStoreFromConnection(
		persist.DB,
		"session",
		"/",
		86400,
		[]byte(cookieSeed))

	if err != nil {
		panic(err)
	}
}

// RelayListener Functions for communicating with the relay server
// Recommend firing off these calls in a goroutine
// as they will return their results asyncrounsly.
// so that you don't have to keep them around
type RelayListener interface {
	// Add network to database and relay-based servers
	AddNetwork(n model.Network)

	// Remove a network and *all* of it's encoders from the
	// database and traffic
	RemoveNetwork(id model.Network)

	// Add encoder to it's network
	AddEncoder(enc model.Encoder)

	// Logout encoder
	LogoutEncoder(id model.Encoder)

	// Remove encoder from database and from encoder
	DeleteEncoder(id model.Encoder)

	// Mute a captioner to keep them from being able to
	// send data to the encoders
	MuteCaptioner(id model.CaptionerID)

	// Unmute a captioner, allowing them to send data to encoders
	UnmuteCaptioner(id model.CaptionerID)

	// DisconnectCaptioner forcibly disconnects the specified captioner
	DisconnectCaptioner(id model.CaptionerID)

	// Remove a captioner, forcibly disconnecting them
	RemoveCaptioner(id model.CaptionerID)

	// GetConnectedCaptioners gets all of the currently connected
	// captioners for a given network.
	GetConnectedCaptioners(n model.Network) []microphone.CaptionerStatus

	// GetConnectedEncoders get all the currently connected encoders
	GetConnectedEncoders(n model.Network) []model.EncoderID

	// GetListeningNetworks get all of the currently connected and listening networks
	GetListeningNetworks() map[model.NetworkID]bool

	// TryChangeNetworkPort attempts to save a change of port, returning an error if something goes wrong.
	TryChangeNetworkPort(id model.NetworkID, port int) error

	// Change the timeout for this network.
	ChangeNetworkTimeout(id model.NetworkID, seconds int)
}

var relay RelayListener

// Start starts the HTTP server
func Start(l RelayListener) {
	relay = l

	// store.Codecs = securecookie.CodecsFromPairs(securecookie.GenerateRandomKey(32))
	// Configure session cookie options.
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: false,
		Secure:   true,
	}

	// Add all web page routes to the router.
	router := addRoutes()

	// insert middleware for CSRF protection.
	csrfHandle := csrfProtect(router) // func call conditional to the build "prod" tag.

	// Start the HTTP server
	if err := http.ListenAndServe(":4000", csrfHandle); err != nil {
		panic(err)
	}
}

// NetworkPortStateChanged Port listener state changed (Inbound network listener)
func NetworkPortStateChanged(network model.Network, state states.Network) {
	notifyNetworkPortStateChanged(network, state)
}

// CaptionerStateChanged lint
func CaptionerStateChanged(captioner model.CaptionerID, state states.Captioner) {
	notifyCaptionerStateChange(captioner, state)
}

// EncoderStateChanged notify the dashboard that an encoder just changed to a new state.
func EncoderStateChanged(encoder model.Encoder, state states.Encoder) {
	notifyEncoderStateChange(encoder, state)
}

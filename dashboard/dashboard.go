package dashboard

import (
	"net/http"

	"github.com/0x7fffffff/verbatim/model"
	"github.com/gorilla/sessions"

	// Passing lint
	"github.com/0x7fffffff/verbatim/persist"
	// "github.com/gorilla/csrf"
	"github.com/michaeljs1990/sqlitestore"
)

var store *sqlitestore.SqliteStore

func init() {
	var err error
	store, err = sqlitestore.NewSqliteStoreFromConnection(persist.DB, "session", "/", 86400, []byte("7Yw2M)QQ0!7Qz=84BO,4M7eSd'#ZhU"))
	if err != nil {
		panic(err)
	}
}

// Start starts the HTTP server
func Start() {
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
	}

	addRoutes()

	// Switch these lines for production
	// protected := csrf.Protect([]byte("tb82Tg0Hw8vVQ6cO8TP1Yh9D69M0lKX4"))(router)
	// protected := csrf.Protect([]byte("tb82Tg0Hw8vVQ6cO8TP1Yh9D69M0lKX4"), csrf.Secure(false))(router)

	if err := http.ListenAndServe("127.0.0.1:4000", nil /*protected*/); err != nil {
		panic(err)
	}
}

// EncoderState EncoderState
type EncoderState int

const (
	// Connected Connected
	Connected EncoderState = iota
	// Disconnected Disconnected
	Disconnected
)

// EncoderStateChanged notify the dashboard that an encoder just changed to a new state.
func EncoderStateChanged(encoder model.Encoder, state EncoderState) {

}

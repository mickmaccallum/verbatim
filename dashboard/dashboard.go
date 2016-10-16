package dashboard

import (
	"net/http"

	// Passing lint
	_ "github.com/0x7fffffff/verbatim/persist"
	"github.com/0x7fffffff/verbatim/websocket"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
)

// Start starts the HTTP server
func Start() {
	router := mux.NewRouter()
	addRoutes(router)
	websocket.Start(router)

	// Switch these lines for production
	// protected := csrf.Protect([]byte("tb82Tg0Hw8vVQ6cO8TP1Yh9D69M0lKX4"))(router)
	protected := csrf.Protect([]byte("tb82Tg0Hw8vVQ6cO8TP1Yh9D69M0lKX4"), csrf.Secure(false))(router)

	if err := http.ListenAndServe("127.0.0.1:4000", protected); err != nil {
		panic(err)
	}
}

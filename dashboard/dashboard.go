package dashboard

import (
	"net/http"
	"time"

	// Passing lint
	_ "github.com/0x7fffffff/verbatim/persist"
	"github.com/gorilla/mux"
)

// Start starts the HTTP server
func Start() {
	router := mux.NewRouter()

	addRoutes(router)

	srv := &http.Server{
		Handler:      router,
		Addr:         "127.0.0.1:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		panic(err)
	}
}

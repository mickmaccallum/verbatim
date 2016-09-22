package dashboard

import (
	"log"
	"net/http"
)

// Start starts the HTTP server
func Start() {
	addRoutes()

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	} else {
		log.Println("Server running on :8080")
	}
}

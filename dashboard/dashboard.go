package dashboard

import (
	"log"
	"net/http"

	_ "github.com/0x7fffffff/verbatim/persist"
	// Passing lint
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

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

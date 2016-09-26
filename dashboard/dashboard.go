package dashboard

import (
	"log"
	"net/http"

	// Passing lint
	_ "github.com/0x7fffffff/verbatim/persist"
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

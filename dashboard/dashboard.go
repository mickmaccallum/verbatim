package dashboard

import (
	"database/sql"
	"log"
	"net/http"

	// Passing lint
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("sqlite3", "database.db")
	checkErr(err)
}

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

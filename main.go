package main

import (
	"database/sql"
	"io/ioutil"

	"github.com/0x7fffffff/verbatim/dashboard"
	_ "github.com/0x7fffffff/verbatim/relay"

	_ "github.com/mattn/go-sqlite3"
)

func configureDatabase(database *sql.DB) (sql.Result, error) {
	bytes, err := ioutil.ReadFile("sql/create_tables.sql")
	checkErr(err)

	return database.Exec(string(bytes))
}

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("sqlite3", "database.db")
	checkErr(err)
	err = db.Ping()
	checkErr(err)

	_, err = configureDatabase(db)

	checkErr(err)
}

func main() {
	dashboard.Start()
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

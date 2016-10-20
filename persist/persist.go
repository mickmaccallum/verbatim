package persist

import (
	"database/sql"
	"io/ioutil"

	// Linter
	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func init() {
	var err error
	DB, err = sql.Open("sqlite3", "database.db")
	checkErr(err)
	err = DB.Ping()
	checkErr(err)

	_, err = configureDatabase(DB)
	checkErr(err)
}

func configureDatabase(database *sql.DB) (sql.Result, error) {
	bytes, err := ioutil.ReadFile("sql/create_tables.sql")
	checkErr(err)

	return database.Exec(string(bytes))
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

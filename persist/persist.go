package persist

import (
	"database/sql"
	"io/ioutil"
)

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

package persist

import (
	"database/sql"
	// "io/ioutil"

	// Linter
	_ "github.com/mattn/go-sqlite3"
)

// DB lint
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
	ddl := `
		create table if not exists admin (
		  id integer primary key,
		  handle text unique not null,
		  hashed_password text not null
		);

		create table if not exists network (
		  id integer primary key,
		  listening_port integer unique not null,
		  name text not null,
		  timeout integer not null
		);

		create table if not exists encoder (
		  id integer primary key,
		  ip_address text not null,
		  port integer not null default(23),
		  name text null default ('New Encoder'),
		  handle text not null,
		  password text not null,
		  network_id integer not null,
		  foreign key(network_id) references network(id)
		);
	`

	// bytes, err := ioutil.ReadFile("sql/create_tables.sql")
	// checkErr(err)

	return database.Exec(ddl)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

func templateOnBase(path string) *template.Template {
	template := template.Must(template.ParseFiles(
		"templates/_base.html",
		path,
	))

	return template
}

func serveStaticFolder(folder string) {
	static := fmt.Sprintf("static%s", folder)
	http.Handle(folder, http.StripPrefix(folder, http.FileServer(http.Dir(static))))
}

func serveRoute(route string, templateName string) {
	http.HandleFunc(route, func(writer http.ResponseWriter, _request *http.Request) {
		template := templateOnBase(fmt.Sprintf("templates/%s", templateName))

		if err := template.Execute(writer, nil); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
	})
}

func setsUpTheTableThing() {
	// bytes, err := ioutil.ReadFile("sql/create_tables.sqlite3")
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS test_table_the_second (`username` VARCHAR, `password` VARCHAR);")
	checkErr(err)
}

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("sqlite3", "database.db")
	checkErr(err)
}

func main() {
	serveStaticFolder("/css/")
	serveStaticFolder("/fonts/")
	serveStaticFolder("/js/")

	serveRoute("/", "_dashboard.html")
	serveRoute("/network.html", "_network.html")

	// stmt, err := db.Prepare("INSERT INTO test_table(username, password) values(?, ?)")
	// checkErr(err)
	//
	// res, err := stmt.Exec("John", "sk2zrule")
	// checkErr(err)
	//
	// id, err := res.LastInsertId()
	// checkErr(err)
	//
	// log.Println(id)

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

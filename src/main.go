package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

var baseTemplate = "templates/_base.html"

func templateOnBase(path string) *template.Template {
	template := template.Must(template.ParseFiles(
		baseTemplate,
		path,
	))

	return template
}

func serveStaticFolder(folder string) {
	static := fmt.Sprintf("static%s", folder)
	http.Handle(folder, http.StripPrefix(folder, http.FileServer(http.Dir(static))))
}

func main() {
	serveStaticFolder("/css/")
	serveStaticFolder("/fonts/")
	serveStaticFolder("/js/")

	http.HandleFunc("/", func(writer http.ResponseWriter, _request *http.Request) {
		template := templateOnBase("templates/_dashboard.html")

		if err := template.Execute(writer, nil); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	} else {
		log.Println("Server running on :8080")
	}
}

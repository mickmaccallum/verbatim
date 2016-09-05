package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
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

func main() {
	serveStaticFolder("/css/")
	serveStaticFolder("/fonts/")
	serveStaticFolder("/js/")

	serveRoute("/", "_dashboard.html")
	serveRoute("/network.html", "_network.html")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	} else {
		log.Println("Server running on :8080")
	}
}

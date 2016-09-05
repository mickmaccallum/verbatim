package main

import (
	"html/template"
	"log"
	"net/http"
)

var baseTemplate = "templates/_base.html"

var dashboardTemplate = template.Must(template.ParseFiles(
	baseTemplate,
	"templates/_dashboard.html",
))

func index(w http.ResponseWriter, r *http.Request) {
	if err := dashboardTemplate.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("templates/static/css"))))
	http.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.Dir("templates/static/fonts"))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("templates/static/js"))))

	http.HandleFunc("/", index)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	} else {
		log.Println("Server running on :8080")
	}
}

package dashboard

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

// func serveRoute(route string, templateName string) {
// 	http.HandleFunc(route, func(writer http.ResponseWriter, _request *http.Request) {
// 		template := templateOnBase(fmt.Sprintf("templates/%s", templateName))
//
// 		if err := template.Execute(writer, nil); err != nil {
// 			http.Error(writer, err.Error(), http.StatusInternalServerError)
// 		}
// 	})
// }

func handleNetworks() {
	http.HandleFunc("/network.html", func(writer http.ResponseWriter, request *http.Request) {
		template := templateOnBase(fmt.Sprintf("templates/_network.html"))
		log.Println(request.URL.Query())
		getNetworks()

		if err := template.Execute(writer, nil); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
	})
}

func handleDashboard() {
	http.HandleFunc("/", func(writer http.ResponseWriter, _request *http.Request) {
		template := templateOnBase(fmt.Sprintf("templates/_dashboard.html"))

		m := make(map[string]int)
		m["route"] = 66

		if err := template.Execute(writer, m); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
	})
}

func addRoutes() {
	// TODO: Guard around admin privileges

	serveStaticFolder("/css/")
	serveStaticFolder("/fonts/")
	serveStaticFolder("/js/")

	handleDashboard()
	handleNetworks()
}

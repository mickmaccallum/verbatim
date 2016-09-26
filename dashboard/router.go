package dashboard

import (
	"errors"
	"github.com/0x7fffffff/verbatim/persist"
	"html/template"
	"net/http"
	"strconv"
)

func templateOnBase(path string) *template.Template {
	template := template.Must(template.ParseFiles(
		"templates/_base.html",
		path,
	))

	return template
}

func serveStaticFolder(folder string) {
	static := "static" + folder

	http.Handle(folder, http.StripPrefix(folder, http.FileServer(http.Dir(static))))
}

func handleNetworks() {
	http.HandleFunc("/network.html", func(writer http.ResponseWriter, request *http.Request) {

		ids := request.URL.Query()["network"]
		if ids == nil {
			clientError(writer, errors.New("Missing network parameter"))
			return
		}

		idString := ids[0]
		if idString == "" {
			clientError(writer, errors.New("Missing network identifier"))
			return
		}

		id, err := strconv.Atoi(idString)
		if err != nil {
			clientError(writer, err)
			return
		}

		network, err := persist.GetNetwork(id)
		if err != nil {
			clientError(writer, err)
			return
		}

		encoders, err := persist.GetEncodersForNetwork(*network)
		if err != nil {
			serverError(writer, err)
		}

		template := templateOnBase("templates/_network.html")
		data := struct {
			Network  persist.Network // Yikes
			Encoders []persist.Encoder
		}{
			*network,
			encoders,
		}

		if err := template.Execute(writer, data); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
	})
}

func handleDashboard() {
	http.HandleFunc("/", func(writer http.ResponseWriter, _request *http.Request) {
		networks, err := persist.GetNetworks()

		if err != nil {
			serverError(writer, err)
			return
		}

		data := struct {
			Networks []persist.Network
		}{
			networks,
		}

		template := templateOnBase("templates/_dashboard.html")
		if err = template.Execute(writer, data); err != nil {
			serverError(writer, err)
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

func clientError(writer http.ResponseWriter, err error) {
	http.Error(writer, err.Error(), http.StatusBadRequest)
}

func serverError(writer http.ResponseWriter, err error) {
	http.Error(writer, err.Error(), http.StatusInternalServerError)
}

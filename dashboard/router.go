package dashboard

import (
	"errors"
	"html/template"
	"net/http"
	"strconv"

	"github.com/0x7fffffff/verbatim/persist"
	"github.com/gorilla/mux"
)

func templateOnBase(path string) *template.Template {
	template := template.Must(template.ParseFiles(
		"templates/_base.html",
		path,
	))

	return template
}

func serveStaticFolder(folder string, router *mux.Router) {
	static := "static" + folder

	http.Handle(folder, http.StripPrefix(folder, http.FileServer(http.Dir(static))))
}

func handleNetworksPage(router *mux.Router) {
	router.HandleFunc("/encoder/add", func(writer http.ResponseWriter, request *http.Request) {

		if isMethodNotAllowed("POST", writer, request) {
			return
		}

		// decoder := json.NewDecoder(request.Body)

		log.Println(request.Body)
	})

	router.HandleFunc("/network.html", func(writer http.ResponseWriter, request *http.Request) {

		if request.Method != "GET" {
			http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

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

func handleDashboardPage(router *mux.Router) {
	router.HandleFunc("/", func(writer http.ResponseWriter, _request *http.Request) {
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

	router := mux.NewRouter()

	serveStaticFolder("/css/", router)
	serveStaticFolder("/fonts/", router)
	serveStaticFolder("/js/", router)

	handleDashboardPage(router)
	handleNetworksPage(router)

	http.Handle("/", router)
}

func clientError(writer http.ResponseWriter, err error) {
	http.Error(writer, err.Error(), http.StatusBadRequest)
}

func serverError(writer http.ResponseWriter, err error) {
	http.Error(writer, err.Error(), http.StatusInternalServerError)
}

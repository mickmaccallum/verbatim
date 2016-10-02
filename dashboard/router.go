package dashboard

import (
	"database/sql"
	"errors"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/0x7fffffff/verbatim/persist"
	// Pin in this
	_ "github.com/gorilla/csrf"
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
	fileServer := http.FileServer(http.Dir(static))
	router.PathPrefix(folder).Handler(http.StripPrefix(folder, fileServer))
}

func handleNetworksPage(router *mux.Router) {
	router.HandleFunc("/encoder/add", func(writer http.ResponseWriter, request *http.Request) {
		log.Println(request.Cookies())
		cookie, err := request.Cookie("current_network_id")
		if err != nil {
			log.Println(err.Error())
			clientError(writer, err)
			return
		}

		networkIDString := cookie.Value
		ip, portString, name :=
			request.FormValue("ip"),
			request.FormValue("port"),
			request.FormValue("name")

		if len(ip) < 7 || len(ip) > 15 || len(portString) < 1 || len(portString) > 5 || len(networkIDString) == 0 {
			clientError(writer, errors.New("Invalid data"))
			return
		}

		port, err := strconv.Atoi(portString)
		if err != nil {
			log.Println(err.Error())
			clientError(writer, err)
			return
		}

		networkID, err := strconv.Atoi(networkIDString)
		if err != nil {
			log.Println(err.Error())
			clientError(writer, err)
			return
		}

		encoder := persist.Encoder{
			IPAddress: ip,
			Name:      sql.NullString{String: name, Valid: true},
			Port:      port,
			Status:    0,
			NetworkID: networkID,
		}

		var network *persist.Network
		network, err = persist.GetNetwork(networkID)
		if err != nil {
			log.Println(err.Error())
			clientError(writer, err)
			return
		}

		err = persist.AddEncoder(encoder, *network)
		if err != nil {
			log.Println(err.Error())
			serverError(writer, err)
		}

		// fmt.Fprint(writer, "{\"message\":\"got it!\"}")

		writer.WriteHeader(200)

	}).Methods("POST")

	router.HandleFunc("/networks/{network_id:[0-9]+}", func(writer http.ResponseWriter, request *http.Request) {

		vars := mux.Vars(request)
		idString := vars["network_id"]

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
			Network  persist.Network
			Encoders []persist.Encoder
		}{
			*network,
			encoders,
		}

		cookie := &http.Cookie{
			Name:  "current_network_id",
			Value: strconv.Itoa(network.ID),
			Path:  request.URL.Path,
		}
		http.SetCookie(writer, cookie)

		if err := template.Execute(writer, data); err != nil {
			serverError(writer, err)
		}
	}).Methods("GET")
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
	}).Methods("GET")
}

func handleNotFound(writer http.ResponseWriter, request *http.Request) {
	data := struct {
		Location string
	}{
		request.URL.RequestURI(),
	}

	template := templateOnBase("templates/error/_404.html")
	if err := template.Execute(writer, data); err != nil {
		serverError(writer, err)
	}
}

func addRoutes(router *mux.Router) {
	// TODO: Guard around admin privileges

	serveStaticFolder("/css/", router)
	serveStaticFolder("/fonts/", router)
	serveStaticFolder("/js/", router)

	handleDashboardPage(router)
	handleNetworksPage(router)

	router.NotFoundHandler = http.HandlerFunc(handleNotFound)

	http.Handle("/", router)
}

func clientError(writer http.ResponseWriter, err error) {
	http.Error(writer, err.Error(), http.StatusBadRequest)
}

func serverError(writer http.ResponseWriter, err error) {
	http.Error(writer, err.Error(), http.StatusInternalServerError)
}

func isMethodNotAllowed(method string, writer http.ResponseWriter, request *http.Request) bool {
	http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
	return method != request.Method
}

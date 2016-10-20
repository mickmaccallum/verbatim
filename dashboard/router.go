package dashboard

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/0x7fffffff/verbatim/model"
	"github.com/0x7fffffff/verbatim/persist"
	"github.com/0x7fffffff/verbatim/websocket"
	"github.com/gorilla/mux"
)

func templateOnBase(path string) *template.Template {
	funcMap := template.FuncMap{
		"inc": func(i int) int {
			return i + 1
		},
	}

	return template.Must(template.New("_base.html").Funcs(funcMap).ParseFiles(
		"templates/_base.html",
		path,
	))
}

func handleLogin(router *mux.Router) {
	router.HandleFunc("/login", func(writer http.ResponseWriter, request *http.Request) {
		data := struct{}{}

		template := templateOnBase("templates/_login.html")
		if err := template.Execute(writer, data); err != nil {
			serverError(writer, err)
		}
	}).Methods("GET")

	router.HandleFunc("/login", func(writer http.ResponseWriter, request *http.Request) {

	}).Methods("POST")
}

func handleCaptionersPage(router *mux.Router) {
	router.HandleFunc("/captioners", func(writer http.ResponseWriter, request *http.Request) {
		data := struct {
			// Networks []persist.Network
		}{
		// networks,
		}

		template := templateOnBase("templates/_captioners.html")
		if err := template.Execute(writer, data); err != nil {
			serverError(writer, err)
		}
	}).Methods("GET")
}

func handleNetworksPage(router *mux.Router) {
	// Add Encoder
	router.HandleFunc("/encoder/add", func(writer http.ResponseWriter, request *http.Request) {
		if err := request.ParseForm(); err != nil {
			clientError(writer, err)
			return
		}

		encoder, err := model.FormValuesToEncoder(request.Form)
		if err != nil {
			clientError(writer, err)
			return
		}

		var network *model.Network
		network, err = persist.GetNetwork(encoder.NetworkID)
		if err != nil {
			clientError(writer, err)
			return
		}

		newEncoder, err := persist.AddEncoder(*encoder, *network)
		if err != nil {
			serverError(writer, err)
			return
		}

		bytes, err := persist.EncoderToJSON(*newEncoder)
		if err != nil {
			serverError(writer, err)
			return
		}

		fmt.Fprint(writer, template.JSStr(bytes))
	}).Methods("POST")

	// Update Encoder
	router.HandleFunc("/encoder/{encoder_id:[0-9]+}", func(writer http.ResponseWriter, request *http.Request) {

		id := identifierFromRequest("encoder_id", request)
		if id == nil {
			clientError(writer, errors.New("Missing encoder identifier"))
			return
		}

		encoder, err := model.FormValuesToEncoder(request.Form)
		if err != nil {
			clientError(writer, err)
			return
		}

		encoder.ID = *id
		err = persist.UpdateEncoder(*encoder)
		if err != nil {
			serverError(writer, err)
			return
		}

		http.Error(writer, "Encoder Updated", http.StatusOK)
	}).Methods("POST")

	// Delete Encoder
	router.HandleFunc("/encoder/{encoder_id:[0-9]+}", func(writer http.ResponseWriter, request *http.Request) {

		id := identifierFromRequest("encoder_id", request)
		if id == nil {
			clientError(writer, errors.New("Missing encoder identifier"))
			return
		}

		encoder, err := persist.GetEncoder(*id)
		if err != nil {
			clientError(writer, err)
			return
		}

		err = persist.DeleteEncoder(*encoder)
		if err != nil {
			serverError(writer, err)
			return
		}

		writer.WriteHeader(http.StatusOK)
	}).Methods("DELETE")

	// Get Encoder
	router.HandleFunc("/networks/{network_id:[0-9]+}", func(writer http.ResponseWriter, request *http.Request) {

		id := identifierFromRequest("network_id", request)
		if id == nil {
			clientError(writer, errors.New("Missing network identifier"))
			return
		}

		network, err := persist.GetNetwork(*id)
		if err != nil {
			clientError(writer, err)
			return
		}

		encoders, err := persist.GetEncodersForNetwork(*network)
		if err != nil {
			serverError(writer, err)
			return
		}

		// t.ExecuteTemplate(w, "signup_form.tmpl", map[string]interface{}{
		// 	csrf.TemplateTag: csrf.TemplateField(r),
		// })
		// csrf.TemplateTag.

		// csrf.TemplateField(r)

		template := templateOnBase("templates/_network.html")
		data := struct {
			Network  model.Network
			Encoders []model.Encoder
			// TemplateTag template.HTML
		}{
			*network,
			encoders,
			// csrf.TemplateField(request),
		}

		if err := template.Execute(writer, data); err != nil {
			serverError(writer, err)
		}
	}).Methods("GET")
}

func handleDashboardPage(router *mux.Router) {
	// Get Dashboard
	router.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		networks, err := persist.GetNetworks()

		if err != nil {
			serverError(writer, err)
			return
		}

		data := struct {
			Networks  []model.Network
			SocketURL string
		}{
			networks,
			"ws://" + request.Host + "/socket", // TODO: Update to wss:// once SSL support is added.
		}

		template := templateOnBase("templates/_dashboard.html")
		if err = template.Execute(writer, data); err != nil {
			serverError(writer, err)
		}
	}).Methods("GET")

	// Add Network
	router.HandleFunc("/network/add", func(writer http.ResponseWriter, request *http.Request) {
		if err := request.ParseForm(); err != nil {
			clientError(writer, err)
			return
		}

		network, err := model.FormValuesToNetwork(request.Form)
		if err != nil {
			clientError(writer, err)
			return
		}

		newNetwork, err := persist.AddNetwork(*network)
		if err != nil {
			serverError(writer, err)
			return
		}

		bytes, err := persist.NetworkToJSON(*newNetwork)
		if err != nil {
			serverError(writer, err)
			return
		}

		fmt.Fprint(writer, template.JSStr(bytes))
	}).Methods("POST")

	// Update Network
	router.HandleFunc("/network/{id:[0-9]+}", func(writer http.ResponseWriter, request *http.Request) {
		networkID := identifierFromRequest("id", request)
		if networkID == nil {
			clientError(writer, errors.New("Invalid Network ID"))
			return
		}

		network, err := persist.GetNetwork(*networkID)
		log.Println(network)
		log.Println(err)

		// http.Error(writer, err.Error(), http.StatusBadRequest)
		http.Error(writer, "", http.StatusOK)
	}).Methods("POST")

	// Delete Network
	router.HandleFunc("/network/{id:[0-9]+}", func(writer http.ResponseWriter, request *http.Request) {

		networkID := identifierFromRequest("id", request)
		if networkID == nil {
			clientError(writer, errors.New("Invalid Network ID"))
			return
		}

		network, err := persist.GetNetwork(*networkID)
		if err != nil {
			clientError(writer, err)
			return
		}

		err = persist.DeleteNetwork(*network)
		if err != nil {
			serverError(writer, err)
			return
		}

		http.Error(writer, "Deleted Network", http.StatusOK)
	}).Methods("DELETE")
}

func generalNotFound(writer http.ResponseWriter, request *http.Request) {
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

func login(writer http.ResponseWriter, request *http.Request) {

}

func addRoutes() {
	router := mux.NewRouter()

	handleLogin(router)
	handleDashboardPage(router)
	handleNetworksPage(router)
	handleCaptionersPage(router)
	websocket.Start(router)

	serveStaticFolder("/css/", router)
	serveStaticFolder("/js/", router)
	serveStaticFolder("/fonts/", router)

	router.NotFoundHandler = http.HandlerFunc(generalNotFound)
	http.Handle("/", router)
}

func serveStaticFolder(folder string, router *mux.Router) {
	static := "static" + folder
	fileServer := http.FileServer(http.Dir(static))
	router.PathPrefix(folder).Handler(http.StripPrefix(folder, fileServer))
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

func identifierFromRequest(identifier string, request *http.Request) *int {
	vars := mux.Vars(request)
	idString := vars[identifier]

	if idString == "" {
		return nil
	}

	id, err := strconv.Atoi(idString)
	if err != nil {
		return nil
	}

	return &id
}

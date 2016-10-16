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

func serveStaticFolder(folder string, router *mux.Router) {
	static := "static" + folder
	fileServer := http.FileServer(http.Dir(static))
	router.PathPrefix(folder).Handler(http.StripPrefix(folder, fileServer))
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
	router.HandleFunc("/encoder/add", func(writer http.ResponseWriter, request *http.Request) {

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
			"ws://" + request.Host + "/socket", // Update to wss:// once SSL support is added.
		}

		template := templateOnBase("templates/_dashboard.html")
		if err = template.Execute(writer, data); err != nil {
			serverError(writer, err)
		}
	}).Methods("GET")

	router.HandleFunc("/network/{id:[0-9]+}", func(writer http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		networkIDString := vars["id"]

		networkID, err := strconv.Atoi(networkIDString)
		if err != nil {
			clientError(writer, err)
			return
		}

		network, err := persist.GetNetwork(networkID)
		log.Println(network)
		log.Println(err)

		// http.Error(writer, err.Error(), http.StatusBadRequest)
		http.Error(writer, "", http.StatusOK)
	}).Methods("POST")

	router.HandleFunc("/network/{id:[0-9]+}", func(writer http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		networkIDString := vars["id"]

		networkID, err := strconv.Atoi(networkIDString)
		if err != nil {
			clientError(writer, err)
			return
		}

		network, err := persist.GetNetwork(networkID)
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
	handleCaptionersPage(router)

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

package dashboard

import (
	"encoding/json"
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
	"github.com/gorilla/sessions"
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

func fetchAdminForSession(session *sessions.Session) (*model.Admin, error) {
	someID, ok := session.Values["admin"]
	if !ok {
		return nil, errors.New("Invalid Session")
	}

	id, ok := someID.(int)
	if !ok {
		return nil, errors.New("Invalid Session")
	}

	return persist.GetAdminForID(id)
}

func checkSessionValidity(request *http.Request) bool {
	session, err := store.Get(request, "session")
	if err != nil {
		return false
	}

	return !session.IsNew
}

func redirectLogin(writer http.ResponseWriter, request *http.Request) {
	http.Redirect(writer, request, "/login", http.StatusSeeOther)
}

func handleAccounts(router *mux.Router) {
	router.HandleFunc("/account", func(writer http.ResponseWriter, request *http.Request) {
		if !checkSessionValidity(request) {
			redirectLogin(writer, request)
			return
		}

		session, err := store.Get(request, "session")
		if err != nil {
			redirectLogin(writer, request)
			return
		}

		admin, err := fetchAdminForSession(session)
		if err != nil {
			clientError(writer, err)
		}
		data := struct {
			Admin model.Admin
		}{
			*admin,
		}

		template := templateOnBase("templates/_account.html")
		if err := template.Execute(writer, data); err != nil {
			serverError(writer, err)
		}
	}).Methods("GET")

	router.HandleFunc("", func(writer http.ResponseWriter, request *http.Request) {

	})
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
		if err := request.ParseForm(); err != nil {
			clientError(writer, err)
			return
		}

		handles := request.Form["handle"]
		passwords := request.Form["password"]

		if len(handles) != 1 || len(passwords) != 1 {
			redirectLogin(writer, request)
			return
		}

		handle := handles[0]
		password := passwords[0]

		admin, err := persist.GetAdminForCredentials(handle, password)
		if err != nil {
			redirectLogin(writer, request)
			return
		}

		session, err := store.Get(request, "session")
		if err != nil {
			redirectLogin(writer, request)
			return
		}

		session.Values["admin"] = admin.ID

		if err = session.Save(request, writer); err != nil {
			redirectLogin(writer, request)
			return
		}

		http.Redirect(writer, request, "/", http.StatusSeeOther)
	}).Methods("POST")

	router.HandleFunc("/logout", func(writer http.ResponseWriter, request *http.Request) {
		defer redirectLogin(writer, request)

		session, err := store.Get(request, "session")
		if err != nil {
			return
		}

		_ = store.Delete(request, writer, session)
	}).Methods("POST")
}

func handleCaptionersPage(router *mux.Router) {
	router.HandleFunc("/captioners", func(writer http.ResponseWriter, request *http.Request) {
		if !checkSessionValidity(request) {
			redirectLogin(writer, request)
			return
		}

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
		if !checkSessionValidity(request) {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}

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
		network, err = persist.GetNetwork(int(encoder.NetworkID))
		if err != nil {
			clientError(writer, err)
			return
		}

		newEncoder, err := persist.AddEncoder(*encoder, *network)
		if err != nil {
			serverError(writer, err)
			return
		}

		relay.AddEncoder(*newEncoder)

		bytes, err := persist.EncoderToJSON(*newEncoder)
		if err != nil {
			serverError(writer, err)
			return
		}

		fmt.Fprint(writer, string(bytes))
	}).Methods("POST")

	// Update Encoder
	router.HandleFunc("/encoder/{encoder_id:[0-9]+}", func(writer http.ResponseWriter, request *http.Request) {
		if !checkSessionValidity(request) {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}

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

		encoder.ID = model.EncoderID(*id)
		err = persist.UpdateEncoder(*encoder)
		if err != nil {
			serverError(writer, err)
			return
		}

		writer.WriteHeader(http.StatusOK)
	}).Methods("POST")

	// Delete Encoder
	router.HandleFunc("/encoder/{encoder_id:[0-9]+}", func(writer http.ResponseWriter, request *http.Request) {
		if !checkSessionValidity(request) {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}

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

		relay.DeleteEncoder(*encoder)

		writer.WriteHeader(http.StatusOK)
	}).Methods("DELETE")

	// Get Encoder
	router.HandleFunc("/networks/{network_id:[0-9]+}", func(writer http.ResponseWriter, request *http.Request) {
		if !checkSessionValidity(request) {
			redirectLogin(writer, request)
			return
		}

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
			Network   model.Network
			Encoders  []model.Encoder
			SocketURL string
			// TemplateTag template.HTML
		}{
			*network,
			encoders,
			"ws://" + request.Host + "/socket", // TODO: Update to wss:// once SSL support is added.
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
		if !checkSessionValidity(request) {
			redirectLogin(writer, request)
			return
		}

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
		if !checkSessionValidity(request) {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}

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

		relay.AddNetwork(*newNetwork)

		bytes, err := persist.NetworkToJSON(*newNetwork)
		if err != nil {
			serverError(writer, err)
			return
		}

		fmt.Fprint(writer, string(bytes))
	}).Methods("POST")

	// Update Network
	router.HandleFunc("/network/{id:[0-9]+}", func(writer http.ResponseWriter, request *http.Request) {
		if !checkSessionValidity(request) {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}

		networkID := identifierFromRequest("id", request)
		if networkID == nil {
			clientError(writer, errors.New("Invalid Network ID"))
			return
		}

		network, err := persist.GetNetwork(*networkID)
		log.Println(network)
		log.Println(err)
		writer.WriteHeader(http.StatusOK)
	}).Methods("POST")

	// Delete Network
	router.HandleFunc("/network/{id:[0-9]+}", func(writer http.ResponseWriter, request *http.Request) {
		if !checkSessionValidity(request) {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}

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

		relay.RemoveNetwork(*network)
		writer.WriteHeader(http.StatusOK)
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

func handleJohnEchoRoute(router *mux.Router) {
	router.HandleFunc("/john/echo", func(writer http.ResponseWriter, request *http.Request) {
		if err := request.ParseForm(); err != nil {
			clientError(writer, err)
			return
		}

		bytes, err := json.Marshal(request.Form)
		if err != nil {
			serverError(writer, err)
			return
		}

		fmt.Fprint(writer, string(bytes))
	}).Methods("POST")
}

func addRoutes() *mux.Router {
	router := mux.NewRouter()

	handleLogin(router)
	handleDashboardPage(router)
	handleNetworksPage(router)
	handleCaptionersPage(router)
	handleJohnEchoRoute(router)
	handleAccounts(router)

	serveStaticFolder("/css/", router)
	serveStaticFolder("/js/", router)
	serveStaticFolder("/fonts/", router)
	serveStaticFolder("/json/", router)

	websocket.Start(router)

	router.NotFoundHandler = http.HandlerFunc(generalNotFound)
	http.Handle("/", router)

	return router
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

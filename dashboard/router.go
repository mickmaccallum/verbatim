package dashboard

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"strconv"

	"golang.org/x/crypto/bcrypt"

	"github.com/0x7fffffff/verbatim/dashboard/websocket"
	"github.com/0x7fffffff/verbatim/model"
	"github.com/0x7fffffff/verbatim/persist"
	"github.com/0x7fffffff/verbatim/states"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

// creates the base template into which all subtemplates for individual
// pages will be rendered.
func templateOnBase(path string) *template.Template {
	funcMap := template.FuncMap{
		"inc": func(i int) int {
			return i + 1
		},
		"simplePlural": func(word string, count int) string {
			if count == 1 {
				return word
			}

			return word + "s"
		},
		"captionerStatus": func(status states.Captioner) string {
			switch status {
			case 0:
				return "Connected"
			case 1:
				return "Disconnecting"
			case 2:
				return "Muted"
			case 3:
				return "Unmuted"
			default:
				return "Disconnected"
			}
		},
		"networkStatus": func(status states.Network) string {
			switch status {
			case 0:
				return "Connected"
			case 1:
				return "Listening"
			case 2:
				return "Listening Failed"
			case 3:
				return "Closed"
			case 4:
				return "Deleted"
			default:
				return "Disconnected"
			}
		},
		"encoderStatus": func(status states.Encoder) string {
			switch status {
			case 0:
				return "Connected"
			case 1:
				return "Connecting"
			case 2:
				return "Authentication Failed"
			case 3:
				return "Writes Failing"
			default:
				return "Disconnected"

			}
		},
		// Removes current admin from list of admins.
		"filterAdmin": func(admin model.Admin, admins []model.Admin) []model.Admin {
			var filteredAdmins []model.Admin
			for _, value := range admins {
				if value != admin {
					filteredAdmins = append(filteredAdmins, value)
				}
			}

			return filteredAdmins
		},
	}

	return template.Must(template.New("_base.html").Funcs(funcMap).ParseFiles(
		"templates/_base.html",
		path,
	))
}

// creates the base params that will be passed to all templates when
// they are rendered.
func templateParamsOnBase(new map[string]interface{}, request *http.Request) map[string]interface{} {
	session, err := store.Get(request, "session")
	var showAccount bool
	if err == nil {
		showAccount = !session.IsNew
	} else {
		showAccount = false
	}

	base := map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(request),
		"SocketURL":      "ws://" + request.Host + "/socket", // TODO: Update to wss:// once SSL support is added.
		"ShowAccount":    showAccount,
	}

	for k, v := range base {
		new[k] = v
	}

	return new
}

// gets the admin who owns the session associated with a given request.
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

// determines whether or not the session attached to a given request
// is valid.
func checkSessionValidity(request *http.Request) (*sessions.Session, bool) {
	session, err := store.Get(request, "session")
	if err != nil {
		return nil, false
	}

	renewSession(session)

	return session, !session.IsNew
}

// experimental
func renewSession(session *sessions.Session) {
	session.Options.MaxAge = 86400
}

// forcibly redirects the request to the login screen. Use when the user's
// session is determined to be invalid. Redirects to the registration page
// instead of the login page if it is determined that no admins exists yet.
func redirectLogin(writer http.ResponseWriter, request *http.Request) {
	admins, err := persist.GetAdmins()
	if err != nil {

	}

	if len(admins) > 0 {
		http.Redirect(writer, request, "/login", http.StatusSeeOther)
		return
	}

	host, _, err := net.SplitHostPort(request.RemoteAddr)
	if err != nil {

	}

	ip := net.ParseIP(host)
	if ip.IsLoopback() {
		http.Redirect(writer, request, "/register", http.StatusSeeOther)
	} else {
		http.NotFound(writer, request)
	}
}

// Handles all routes related to the account page.
func handleAccountsPage(router *mux.Router) {
	router.HandleFunc("/account", func(writer http.ResponseWriter, request *http.Request) {
		session, sessionOk := checkSessionValidity(request)
		if !sessionOk {
			redirectLogin(writer, request)
			return
		}

		admin, err := fetchAdminForSession(session)
		if err != nil {
			clientError(writer, err)
			return
		}

		admins, err := persist.GetAdmins()
		if err != nil {
			serverError(writer, err)
			return
		}

		data := map[string]interface{}{
			"Admin":  *admin,
			"Admins": admins,
		}

		template := templateOnBase("templates/_account.html")
		if err = template.Execute(writer, templateParamsOnBase(data, request)); err != nil {
			serverError(writer, err)
		}
	}).Methods("GET")

	router.HandleFunc("/account/add", func(writer http.ResponseWriter, request *http.Request) {
		_, sessionOk := checkSessionValidity(request)
		if !sessionOk {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}

		if err := request.ParseForm(); err != nil {
			clientError(writer, err)
			return
		}

		admin, err := model.FormValuesToAdmin(request.Form)
		if err != nil {
			clientError(writer, err)
			return
		}

		finalAdmin, err := persist.AddAdmin(*admin)
		if err != nil {
			serverError(writer, err)
			return
		}

		bytes, err := json.Marshal(finalAdmin)
		if err != nil {
			serverError(writer, err)
			return
		}

		fmt.Fprint(writer, string(bytes))
	}).Methods("POST")

	// Delete admin account
	router.HandleFunc("/account/delete/{admin_id:[0-9]+}", func(writer http.ResponseWriter, request *http.Request) {
		_, sessionOk := checkSessionValidity(request)
		if !sessionOk {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}

		id := identifierFromRequest("admin_id", request)
		if id == nil {
			clientError(writer, errors.New("Missing admin identifier"))
			return
		}

		admin, err := persist.GetAdminForID(*id)
		if err != nil {
			clientError(writer, err)
			return
		}

		err = persist.DeleteAdmin(*admin)
		if err != nil {
			serverError(writer, err)
			return
		}

		writer.WriteHeader(http.StatusOK)
	}).Methods("POST")

	router.HandleFunc("/account/handle", func(writer http.ResponseWriter, request *http.Request) {
		session, sessionOk := checkSessionValidity(request)
		if !sessionOk {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}

		if err := request.ParseForm(); err != nil {
			clientError(writer, err)
			return
		}

		handle := request.Form.Get("handle")
		if len(handle) == 0 || len(handle) > 255 {
			writer.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		adminID := session.Values["admin"].(int)
		admin, err := persist.GetAdminForID(adminID)
		if err != nil {
			clientError(writer, err)
			return
		}

		admin.Handle = handle
		err = persist.UpdateAdminHandle(*admin)
		if err != nil {
			serverError(writer, err)
			return
		}

		writer.WriteHeader(http.StatusOK)
	}).Methods("POST")

	router.HandleFunc("/account/password", func(writer http.ResponseWriter, request *http.Request) {
		session, sessionOk := checkSessionValidity(request)
		if !sessionOk {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}

		if err := request.ParseForm(); err != nil {
			clientError(writer, err)
			return
		}

		adminID := session.Values["admin"].(int)
		admin, err := persist.GetAdminForID(adminID)
		if err != nil {
			clientError(writer, err)
			return
		}

		oldPassword := request.Form.Get("old_password")
		if len(oldPassword) == 0 {
			clientError(writer, errors.New("Missing old password"))
			return
		}

		if !admin.HasPassword(oldPassword) {
			clientError(writer, errors.New("Old password does not match admin"))
			return
		}

		password, confirmPassword := request.Form.Get("new_password"), request.Form.Get("confirm_new_password")
		if password != confirmPassword {
			clientError(writer, errors.New("Passwords don't match"))
			return
		}

		if len(password) == 0 || len(password) > 255 {
			writer.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			serverError(writer, err)
			return
		}

		admin.HashedPassword = string(hashed)

		err = persist.UpdateAdminPassword(*admin)
		if err != nil {
			serverError(writer, err)
			return
		}

		writer.WriteHeader(http.StatusOK)
	}).Methods("POST")
}

// Handles all routes related to the login page.
func handleLogin(router *mux.Router) {
	router.HandleFunc("/register", func(writer http.ResponseWriter, request *http.Request) {
		data := map[string]interface{}{}

		template := templateOnBase("templates/_registration.html")
		if err := template.Execute(writer, templateParamsOnBase(data, request)); err != nil {
			serverError(writer, err)
		}
	}).Methods("GET")

	router.HandleFunc("/register", func(writer http.ResponseWriter, request *http.Request) {
		if err := request.ParseForm(); err != nil {
			clientError(writer, err)
			return
		}

		admin, err := model.FormValuesToAdmin(request.Form)
		if err != nil {
			clientError(writer, err)
			return
		}

		newAdmin, err := persist.AddAdmin(*admin)
		if err != nil {
			serverError(writer, err)
			return
		}

		session, err := store.Get(request, "session")
		if err != nil {
			serverError(writer, err)
			return
		}

		session.Values["admin"] = newAdmin.ID

		if err = session.Save(request, writer); err != nil {
			serverError(writer, err)
			return
		}

		http.Redirect(writer, request, "/", http.StatusSeeOther)
	}).Methods("POST")

	router.HandleFunc("/login", func(writer http.ResponseWriter, request *http.Request) {
		data := map[string]interface{}{}

		template := templateOnBase("templates/_login.html")
		if err := template.Execute(writer, templateParamsOnBase(data, request)); err != nil {
			serverError(writer, err)
		}
	}).Methods("GET")

	router.HandleFunc("/login", func(writer http.ResponseWriter, request *http.Request) {
		// session, sessionOk := checkSessionValidity(request)
		// if !sessionOk {
		// 	writer.WriteHeader(http.StatusUnauthorized)
		// 	return
		// }

		if err := request.ParseForm(); err != nil {
			clientError(writer, err)
			return
		}

		handles := request.Form["handle"]
		passwords := request.Form["password"]

		if len(handles) != 1 || len(passwords) != 1 {
			log.Println("incorrect length")
			redirectLogin(writer, request)
			return
		}

		handle := handles[0]
		password := passwords[0]

		admin, err := persist.GetAdminForCredentials(handle, password)
		if err != nil {
			log.Println("failed to lookup admin with credentials")
			redirectLogin(writer, request)
			return
		}

		session, err := store.Get(request, "session")
		if err != nil {
			log.Println(err.Error())
			log.Println("couldn't get session")
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
		session, sessionOk := checkSessionValidity(request)
		if !sessionOk {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}

		err := store.Delete(request, writer, session)
		if err != nil {
			log.Println(err.Error())
		}

		redirectLogin(writer, request)
	}).Methods("POST")
}

// Handles all routes related to the networks page.
func handleNetworksPage(router *mux.Router) {
	// Add Encoder
	router.HandleFunc("/encoder/add", func(writer http.ResponseWriter, request *http.Request) {
		_, sessionOk := checkSessionValidity(request)
		if !sessionOk {
			log.Println("session not okay")
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
		_, sessionOk := checkSessionValidity(request)
		if !sessionOk {
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
	router.HandleFunc("/encoder/delete/{encoder_id:[0-9]+}", func(writer http.ResponseWriter, request *http.Request) {
		_, sessionOk := checkSessionValidity(request)
		if !sessionOk {
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
	}).Methods("POST")

	// Get Network
	router.HandleFunc("/network/{network_id:[0-9]+}", func(writer http.ResponseWriter, request *http.Request) {
		_, sessionOk := checkSessionValidity(request)
		if !sessionOk {
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

		connectedEncoders := relay.GetConnectedEncoders(*network)
		for _, encoder := range encoders {
			for _, connectedEncoderID := range connectedEncoders {
				if encoder.ID == connectedEncoderID {
					encoder.Status = states.EncoderConnected
					break
				}
			}
		}

		captioners := relay.GetConnectedCaptioners(*network)
		data := map[string]interface{}{
			"Network":    *network,
			"Encoders":   encoders,
			"Captioners": captioners,
		}

		template := templateOnBase("templates/_network.html")
		if err = template.Execute(writer, templateParamsOnBase(data, request)); err != nil {
			serverError(writer, err)
		}
	}).Methods("GET")

	router.HandleFunc("/captioners/mute", func(writer http.ResponseWriter, request *http.Request) {
		_, sessionOk := checkSessionValidity(request)
		if !sessionOk {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}

		captioner, err := model.FormValuesToCaptionerID(request.Form)
		if err != nil {
			clientError(writer, err)
			return
		}

		relay.MuteCaptioner(*captioner)
		writer.WriteHeader(http.StatusOK)
	}).Methods("POST")

	router.HandleFunc("/captioners/unmute", func(writer http.ResponseWriter, request *http.Request) {
		_, sessionOk := checkSessionValidity(request)
		if !sessionOk {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}

		captioner, err := model.FormValuesToCaptionerID(request.Form)
		if err != nil {
			clientError(writer, err)
			return
		}

		relay.UnmuteCaptioner(*captioner)
		writer.WriteHeader(http.StatusOK)
	}).Methods("POST")

	router.HandleFunc("/captioner/disconnect", func(writer http.ResponseWriter, request *http.Request) {
		_, sessionOk := checkSessionValidity(request)
		if !sessionOk {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}

		captioner, err := model.FormValuesToCaptionerID(request.Form)
		if err != nil {
			clientError(writer, err)
			return
		}

		relay.DisconnectCaptioner(*captioner)
		writer.WriteHeader(http.StatusOK)
	}).Methods("POST")
}

// Handles all routes related to the dashboard page.
func handleDashboardPage(router *mux.Router) {
	// Get Dashboard
	router.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		_, sessionOk := checkSessionValidity(request)
		if !sessionOk {
			redirectLogin(writer, request)
			return
		}

		networks, err := persist.GetNetworks()
		if err != nil {
			serverError(writer, err)
			return
		}

		connectedNetworks := relay.GetListeningNetworks()
		for _, network := range networks {
			if connectedNetworks[network.ID] {
				network.State = states.NetworkListening
			}
		}

		data := map[string]interface{}{
			"Networks": networks,
		}

		template := templateOnBase("templates/_dashboard.html")
		if err = template.Execute(writer, templateParamsOnBase(data, request)); err != nil {
			serverError(writer, err)
		}
	}).Methods("GET")

	// Add Network
	router.HandleFunc("/network/add", func(writer http.ResponseWriter, request *http.Request) {
		_, sessionOk := checkSessionValidity(request)
		if !sessionOk {
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
		_, sessionOk := checkSessionValidity(request)
		if !sessionOk {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}

		networkID := identifierFromRequest("id", request)
		if networkID == nil {
			clientError(writer, errors.New("Invalid Network ID"))
			return
		}

		hitNetwork, err := persist.GetNetwork(*networkID)
		if err != nil {
			clientError(writer, errors.New("The specified network does not exist."))
			return
		}

		network, err := model.FormValuesToNetwork(request.Form)
		if err != nil {
			clientError(writer, err)
			return
		}

		network.ID = hitNetwork.ID

		err = persist.UpdateNetwork(*network)
		if err != nil {
			serverError(writer, err)
			return
		}

		if network.Timeout != hitNetwork.Timeout {
			relay.ChangeNetworkTimeout(network.ID, network.Timeout)
		}

		if network.ListeningPort != hitNetwork.ListeningPort {
			err = relay.TryChangeNetworkPort(network.ID, network.ListeningPort)
			if err != nil {
				serverError(writer, err)
				return
			}
		}

		writer.WriteHeader(http.StatusOK)
	}).Methods("POST")

	// Delete Network
	router.HandleFunc("/network/delete/{id:[0-9]+}", func(writer http.ResponseWriter, request *http.Request) {
		_, sessionOk := checkSessionValidity(request)
		if !sessionOk {
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
	}).Methods("POST")
}

// renders a customer 404 page.
func generalNotFound(writer http.ResponseWriter, request *http.Request) {
	data := map[string]interface{}{
		"Location": request.URL.RequestURI(),
	}

	template := templateOnBase("templates/error/_404.html")
	if err := template.Execute(writer, templateParamsOnBase(data, request)); err != nil {
		serverError(writer, err)
	}
}

// adds all the routes to the router.
func addRoutes() *mux.Router {
	router := mux.NewRouter()

	handleLogin(router)
	handleDashboardPage(router)
	handleNetworksPage(router)
	handleAccountsPage(router)

	serveStaticFolder("/css/", router)
	serveStaticFolder("/js/", router)
	serveStaticFolder("/fonts/", router)

	websocket.Start(router)

	router.NotFoundHandler = http.HandlerFunc(generalNotFound)
	http.Handle("/", router)

	return router
}

// used to server static files, like CSS/JavaScript/fonts/etc.
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

// parses the given identifier out of the request path.
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

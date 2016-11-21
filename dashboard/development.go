// +build !prod

// Contents of this file are only compiled if the "prod" build tag
// is omitted when compiling.

package dashboard

import (
	"log"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
)

const (
	cookieSeed = "4'852b9FtL(_61R!q]La1d_BtEi8(*"
)

func csrfProtect(router *mux.Router) http.Handler {
	log.Println("Running In Development Mode")

	// errHandler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

	// 	sess, err := store.Get(request, "session")
	// 	log.Println(sess.IsNew)
	// 	log.Println(err)
	// 	log.Println(sess.)

	// 	log.Println(csrf.FailureReason(request))
	// })

	return csrf.Protect([]byte("tb82Tg0Hw8vVQ6cO8TP1Yh9D69M0lKX4"), csrf.Secure(false) /*, csrf.ErrorHandler(errHandler)*/)(router)
}

// +build !prod

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
	return csrf.Protect([]byte("tb82Tg0Hw8vVQ6cO8TP1Yh9D69M0lKX4"), csrf.Secure(false))(router)
}

package dashboard

import (
	// "log"
	"net/http"
)

// Authenticate auth
func Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, err := store.Get(request, "session-name")
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		// log.Println(session)

		// http.Error(w, http.StatusText(400), 400)

		next.ServeHTTP(writer, request)
	})
}

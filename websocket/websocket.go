package websocket

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{} // use default options

// Start starts the socket
func Start(router *mux.Router) {
	log.Println("starting socket")
	router.HandleFunc("/socket", openSocket)
}

func openSocket(writer http.ResponseWriter, request *http.Request) {
	conn, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Println(err)
		return
	}

	defer conn.Close()
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}

		log.Printf("recv: %s", message)

		err = conn.WriteMessage(messageType, message)
		if err != nil {
			log.Println(err)
			break
		}
	}
}

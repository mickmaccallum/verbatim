package websocket

import (
	"log"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"
	// "sync/atomic"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var messages = make(chan SocketMessage, 10)

// TODO: Figure out how to remove conns on disconnect.
var sockets = make(map[*websocket.Conn]struct{})
var mutex = &sync.Mutex{}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		components := strings.Split(r.RemoteAddr, ":")
		return components[0] == "127.0.0.1"
	},
}

// Start starts the socket
func Start(router *mux.Router) {
	router.HandleFunc("/socket", openSocket)

	go spin()

	time.AfterFunc(10*time.Second, func() {
		payload := make(map[string]string)
		payload["message"] = "Hello, browser"

		m := &SocketMessage{
			Payload: payload,
		}
		log.Println("Sending Message")
		m.Send()
	})
}

func spin() {
	for {
		select {
		case message := <-messages:
			for conn := range sockets {
				wrapper := struct {
					Payload interface{}
				}{
					message.Payload,
				}

				err := conn.WriteJSON(wrapper)
				if err != nil {
					log.Println(err)
				}
			}
		}

		runtime.Gosched()
	}
}

func openSocket(writer http.ResponseWriter, request *http.Request) {
	conn, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Println(err)
		return
	}

	mutex.Lock()
	sockets[conn] = struct{}{}
	mutex.Unlock()

	log.Println(sockets)

	// defer conn.Close()
	// for {
	// 	messageType, message, err := conn.ReadMessage()
	// 	if err != nil {
	// 		log.Println(err)
	// 		break
	// 	}

	// 	log.Printf("recv: %s", message)

	// 	err = conn.WriteMessage(messageType, message)
	// 	if err != nil {
	// 		log.Println(err)
	// 		break
	// 	}
	// }
}

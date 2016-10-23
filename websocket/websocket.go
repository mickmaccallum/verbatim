package websocket

import (
	"log"
	"net/http"
	"runtime"
	"sync"
	"time"
	// "sync/atomic"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var messages = make(chan SocketMessage, 10)

// TODO: Figure out how to remove conns on disconnect.
var connections = make(map[*websocket.Conn]struct{})
var connMutex = &sync.Mutex{}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// components := strings.Split(r.RemoteAddr, ":")
		// return components[0] == "127.0.0.1"
		return r.Header.Get("Origin") == "http://"+r.Host
	},
}

// Start starts the socket
func Start(router *mux.Router) {
	router.HandleFunc("/socket", openSocket)

	go spin()
	// sendTestMessage()
}

func openSocket(writer http.ResponseWriter, request *http.Request) {
	conn, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Println(err)
		return
	}

	connMutex.Lock()
	connections[conn] = struct{}{}
	connMutex.Unlock()

	log.Println(connections)

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

func spin() {
	for {
		select {
		case message := <-messages:
			sendMessage(message)
		}

		runtime.Gosched()
	}
}

func sendMessage(message SocketMessage) {
	for conn := range connections {
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

func sendTestMessage() {
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

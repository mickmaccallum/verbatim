package websocket

import (
	"log"
	"net/http"
	"runtime"
	"sync"
	"time"

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
}

func openSocket(writer http.ResponseWriter, request *http.Request) {
	conn, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Println(err)
		return
	}

	connectionOpened(conn)

	defer conn.Close()
	for {
		// messageType is either text or binary
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err) || websocket.IsUnexpectedCloseError(err) {
				connectionClosed(conn)
			}

			log.Println(err)
			break
		}

		// Echo
		err = conn.WriteMessage(messageType, message)
		if err != nil {
			log.Println(err)
			break
		}
	}
}

func connectionOpened(conn *websocket.Conn) {
	connMutex.Lock()
	connections[conn] = struct{}{}
	connMutex.Unlock()
}

func connectionClosed(conn *websocket.Conn) {
	connMutex.Lock()
	delete(connections, conn)
	connMutex.Unlock()
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

func sendMessage(message SocketMessage) []error {
	var errors []error

	for conn := range connections {
		wrapper := struct {
			Payload interface{}
		}{
			message.Payload,
		}

		err := conn.WriteJSON(wrapper)
		if err != nil {
			errors = append(errors, err)
		}
	}

	return errors
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

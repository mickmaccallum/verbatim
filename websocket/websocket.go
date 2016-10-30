package websocket

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// NotificationType lint
type NotificationType string

const (
	// EncoderState lint
	EncoderState NotificationType = "encoderState"
	// CaptionerState lint
	CaptionerState NotificationType = "captionerState"
)

var messages = make(chan SocketMessage, 10)
var replies = make(chan reply, 10)

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
		var message SocketMessage

		err := conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsCloseError(err) || websocket.IsUnexpectedCloseError(err) {
				connectionClosed(conn)
			}

			log.Println(err)
			break
		}

		reply{
			conn:    conn,
			message: message,
			response: response{
				Error:   nil,
				Message: "Received Message",
			},
		}.send()
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
			broadcastMessage(message)
		case r := <-replies:
			sendReply(r)
		}
	}
}

func sendReply(r reply) error {
	newMessage := SocketMessage{
		Reference: r.message.Reference,
		Payload:   r.response,
	}

	return sendMessage(newMessage, r.conn)
}

func broadcastMessage(message SocketMessage) []error {
	var errors []error

	for conn := range connections {
		err := sendMessage(message, conn)
		if err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

func sendMessage(message SocketMessage, conn *websocket.Conn) error {
	return conn.WriteJSON(message)
}

func sendTestMessage() {
	time.AfterFunc(10*time.Second, func() {
		payload := map[string]string{
			"message": "Hello, browser!",
		}

		m := &SocketMessage{
			Payload: payload,
		}

		m.Send()
	})
}

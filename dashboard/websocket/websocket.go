package websocket

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// NotificationType use to switch over different types of notifications.
type NotificationType string

const (
	// EncoderState represents an event for an encoder changing state.
	EncoderState NotificationType = "encoderState"
	// CaptionerState represents an event for an captioner changing state.
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

	// Spin and hold the web socket connection open to wait for messages.
	for {
		var message SocketMessage

		err := conn.ReadJSON(&message)
		// If an error message is received, disconnect the connection.
		if err != nil {
			if websocket.IsCloseError(err) || websocket.IsUnexpectedCloseError(err) {
				connectionClosed(conn)
			}

			break
		}

		// Wrap the message in a reply context and sent it back over the connection.
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

// connectionOpened synchronizes adding of a connection to the connection pool
func connectionOpened(conn *websocket.Conn) {
	connMutex.Lock()
	connections[conn] = struct{}{}
	connMutex.Unlock()
}

// connectionClosed synchronizes removing of a connection to the connection pool
func connectionClosed(conn *websocket.Conn) {
	connMutex.Lock()
	delete(connections, conn)
	connMutex.Unlock()
}

// wait for message events, and sent them as they are received.
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

// sendReply sends a message over a connection with a reply context attached.
func sendReply(r reply) error {
	newMessage := SocketMessage{
		Reference: r.message.Reference,
		Payload:   r.response,
	}

	return sendMessage(newMessage, r.conn)
}

// broadcastMessage Sends the given message on every active web socket connection.
func broadcastMessage(message SocketMessage) map[*websocket.Conn]error {
	var errors map[*websocket.Conn]error

	// enumerate all the active connections, and send the message.
	for conn := range connections {
		err := sendMessage(message, conn)
		if err != nil {
			errors[conn] = err
		}
	}

	return errors
}

// sendMessage Writes the given message over the given connection as JSON.
func sendMessage(message SocketMessage, conn *websocket.Conn) error {
	return conn.WriteJSON(message)
}

// sendTestMessage Not actively used. Just here for testing message sends.
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

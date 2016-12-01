package websocket

import (
	"log"
	"net/http"
	"sync"

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
	// NetworkState represents an event for a network changing state
	NetworkState NotificationType = "networkState"
)

var messages = make(chan SocketMessage, 10)
var connections = make(map[*websocket.Conn]chan SocketMessage)
var connMutex = &sync.Mutex{}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
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

	go connectionOpened(conn)

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
		rep := reply{
			conn:    conn,
			message: message,
			response: response{
				Error:   nil,
				Message: "Received Message",
			},
		}

		sendReply(rep)
	}
}

// connectionOpened synchronizes adding of a connection to the connection pool
func connectionOpened(conn *websocket.Conn) {
	inboundMessages := make(chan SocketMessage)

	connMutex.Lock()
	connections[conn] = inboundMessages
	connMutex.Unlock()

	for message := range inboundMessages {
		conn.WriteJSON(message)
	}
}

// connectionClosed synchronizes removing of a connection to the connection pool
func connectionClosed(conn *websocket.Conn) {
	connMutex.Lock()
	close(connections[conn])
	delete(connections, conn)
	connMutex.Unlock()
}

// wait for message events, and sent them as they are received.
func spin() {
	for {
		select {
		case message := <-messages:
			for _, channel := range connections {
				channel <- message
			}
		}
	}
}

// sendReply sends a message over a connection with a reply context attached.
func sendReply(r reply) error {
	newMessage := SocketMessage{
		Reference: r.message.Reference,
		Payload:   r.response,
	}

	return r.conn.WriteJSON(newMessage)
}

// sendTestMessage Not actively used. Just here for testing message sends.
// func sendTestMessage() {
// 	time.AfterFunc(10*time.Second, func() {
// 		payload := map[string]string{
// 			"message": "Hello, browser!",
// 		}

// 		m := &SocketMessage{
// 			Payload: payload,
// 		}

// 		m.Send()
// 	})
// }

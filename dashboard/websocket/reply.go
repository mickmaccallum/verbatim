package websocket

import "github.com/gorilla/websocket"

// reply reply represents a context for a message to be sent in
// response to a message received
type reply struct {
	conn     *websocket.Conn
	message  SocketMessage
	response response
}

// send enqueues the reply to be sent.
func (r reply) send() {
	go func() {
		replies <- r
	}()
}

package websocket

// SocketMessage A message to be emitted over all open websockets
type SocketMessage struct {
	Reference *int
	Payload   interface{}
}

// Send emits the receiver as a message across all websocket connections.
func (message SocketMessage) Send() {
	messages <- message
}

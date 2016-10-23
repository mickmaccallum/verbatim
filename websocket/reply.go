package websocket

import "github.com/gorilla/websocket"

type reply struct {
	conn     *websocket.Conn
	message  SocketMessage
	response response
}

func (r reply) send() {
	go func() {
		replies <- r
	}()
}

package dashboard

import (
	"github.com/0x7fffffff/verbatim/model"
	"github.com/0x7fffffff/verbatim/websocket"
)

func notifyEncoderStateChange(encoder model.Encoder, state EncoderState) {
	message := websocket.SocketMessage{
		Payload: map[websocket.NotificationType]EncoderState{
			websocket.EncoderState: state,
		},
	}

	message.Send()
}

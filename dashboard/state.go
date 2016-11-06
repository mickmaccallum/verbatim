package dashboard

import (
	"github.com/0x7fffffff/verbatim/dashboard/websocket"
	"github.com/0x7fffffff/verbatim/model"
	"github.com/0x7fffffff/verbatim/states"
)

func notifyCaptionerStateChange(captioner model.CaptionerID, state states.Captioner) {
	message := websocket.SocketMessage{
		Payload: map[websocket.NotificationType]interface{}{
			websocket.CaptionerState: map[string]interface{}{
				"state": int(state),
				// "captionerId": captioner
			},
		},
	}

	message.Send()
}

func notifyEncoderStateChange(encoder model.Encoder, state states.Encoder) {
	message := websocket.SocketMessage{
		Payload: map[websocket.NotificationType]interface{}{
			websocket.EncoderState: map[string]interface{}{
				"state":     int(state),
				"encoderId": encoder.ID,
			},
		},
	}

	message.Send()
}

package dashboard

import (
	"github.com/0x7fffffff/verbatim/microphone"
	"github.com/0x7fffffff/verbatim/model"
	"github.com/0x7fffffff/verbatim/states"
	"github.com/0x7fffffff/verbatim/websocket"
)

func notifyCaptionerStateChange(captioner microphone.CaptionerStatus, state states.Captioner) {
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

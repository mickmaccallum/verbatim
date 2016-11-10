package dashboard

import (
	"github.com/0x7fffffff/verbatim/dashboard/websocket"
	"github.com/0x7fffffff/verbatim/model"
	"github.com/0x7fffffff/verbatim/states"
)

func notifyNetworkPortStateChanged(network model.Network, state states.Network) {
	websocket.SocketMessage{
		Payload: wrapState(websocket.NetworkState, state),
	}.Send()
}

func notifyCaptionerStateChange(captioner model.CaptionerID, state states.Captioner) {
	message := websocket.SocketMessage{
		Payload: map[websocket.NotificationType]interface{}{
			websocket.CaptionerState: map[string]interface{}{
				"state":       int(state),
				"captionerId": captioner,
			},
		},
	}

	message.Send()
}

func notifyEncoderStateChange(encoder model.Encoder, state states.Encoder) {
	websocket.SocketMessage{
		Payload: wrapState(websocket.EncoderState, state),
	}.Send()
}

func wrapState(t websocket.NotificationType, s interface{}) websocket.SocketMessage {
	return websocket.SocketMessage{
		Payload: map[websocket.NotificationType]interface{}{
			t: s,
		},
	}
}

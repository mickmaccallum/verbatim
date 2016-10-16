package megaphone

import (
	"github.com/0x7fffffff/verbatim/model"
	"github.com/0x7fffffff/verbatim/persist"
	"log"
)

type MegaphoneListener interface {
	// Logged into encoder properly
	LoginSucceeded(m model.Encoder)
}

func NotifyNetworkAdded(n model.Network) error {

	return nil
}

func NotifyNetworkRemoved(n model.Network) error {

	return nil
}

func NotifyEncoderAdded(m model.Encoder) error {

	return nil
}

func NotifyEncoderRemoved(m model.Encoder) error {

	return nil
}

var l MegaphoneListener

func Start(ml MegaphoneListener) {
	l = ml
	logIntoExistingEncoders()
}

var encodersByNetwork map[model.Network][]model.Encoder
var networksById map[int]model.Network

func logIntoExistingEncoders() {
	networks, err := persist.GetNetworks()
	if err != nil {
		log.Fatal("Unable to connect to database!")
	}
	var networksById = make(map[int]model.Network)
	for _, val := range networks {
		networksById[val.ID] = val
	}

	encoders, err := persist.GetEncoders()
	if err != nil {
		log.Fatal("Unable to connect to database!")
	}
	encodersByNetwork = make(map[model.Network][]model.Encoder)

	for _, val := range encoders {
		var encoderList []model.Encoder
		if encoderList, found := encodersByNetwork[networksById[val.NetworkID]]; found {
			encoderList = append(encoderList, val)
		} else {
			encoderList = []model.Encoder{val}
		}
		encodersByNetwork[networksById[val.NetworkID]] = encoderList
	}
}

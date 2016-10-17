package megaphone

import (
	"github.com/0x7fffffff/verbatim/model"
	"github.com/0x7fffffff/verbatim/persist"
	"log"
	"net"
)

type MegaphoneListener interface {
	// Logged into encoder properly
	LoginSucceeded(enc model.Encoder)
	LoginFailed(enc model.Encoder)
}

func NotifyNetworkAdded(n model.Network) {
	networkAdded <- n
}

func NotifyNetworkRemoved(n model.Network) {
	networkRemoved <- n
}

func NotifyEncoderAdded(enc model.Encoder) {
	encoderAdded <- enc
}

func NotifyEncoderRemoved(enc model.Encoder) {
	encoderRemoved <- enc
}

var l MegaphoneListener

func Start(ml MegaphoneListener) {
	l = ml
	setupEncoders()
}

// Crud notifications
var (
	networkAdded   = make(chan model.Network, 10)
	networkRemoved = make(chan model.Network, 10)
	encoderRemoved = make(chan model.Encoder, 10)
	encoderAdded   = make(chan model.Encoder, 10)
)

type NetworkID int
type EncoderID int

// Send notifications (coming from Relay server)
var sendOnEncoders = make(map[NetworkID][]chan string)

// Maps for tracking state of encoders, and for resolving sense
var (
	encodersByNetwork map[NetworkID][]model.Encoder
	networksById      map[NetworkID]model.Network
)

func GetEncoderState() {

}

func setupEncoders() {
	loadEncoders()
	sendOnEncoders = make(map[NetworkID][]chan string)
	for network, encoders := range encodersByNetwork {
		sendOnEncoders[network] = make([]chan string, len(encoders))
		for idx, enc := range encoders {
			inbound := make(chan string)
			sendOnEncoders[network][idx] = inbound
			go handleEncoder(enc, inbound)
		}
	}
	manageHairyBallOPain()
}

//
func manageHairyBallOPain() {
	for {
		// If any of the channels below are closed, crash the program, something when horribly wrong...
		select {
		case newNet, ok := <-networkAdded:
			if !ok {
				log.Print("Closed network addition channel!")
				return
			}
			encodersByNetwork[NetworkID(newNet.ID)] = make([]model.Encoder, 0)

		case newEnc, ok := <-encoderAdded:
			if !ok {
				log.Print("Closed network addition channel!")
				return
			}
			if encoders, found := encodersByNetwork[NetworkID(newEnc.NetworkID)]; found {
				inbound := make(chan string)
				encoders = append(encoders, newEnc)
			} else {

			}

		}
	}
}
func loadEncoders() {
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
	encodersByNetwork = make(map[NetworkID][]model.Encoder)

	for _, val := range encoders {
		var encoderList []model.Encoder
		if encoderList, found := encodersByNetwork[NetworkID(val.NetworkID)]; found {
			encoderList = append(encoderList, val)
		} else {
			encoderList = []model.Encoder{val}
		}
		encodersByNetwork[NetworkID(val.NetworkID)] = encoderList
	}
}

func SendMessageOnNetwork(id NetworkID) {
}

const LINE_CUT_WIDTH = 32

func writeMessageSegmented(conn net.Conn, msg string) error {
	// Write message in chunks
	for i := 0; i*LINE_CUT_WIDTH < len(msg); i++ {
		begin := i * LINE_CUT_WIDTH
		end := (i + 1) * LINE_CUT_WIDTH
		if end > len(msg) {
			end = len(msg)
		}
		if _, err := conn.Write([]byte(msg[begin:end])); err != nil {
			return err
		}
	}
	return nil
}

func handleEncoder(enc model.Encoder, inbound chan string) {
	conn, err := loginToEncoder(enc)
	if err != nil {
	}
	for {
		select {
		case msg, ok := <-inbound:
			if ok {
				writeMessageSegmented(conn, msg)
			} else {
				//
				return
			}
		}
	}
}

package relay

import (
	"log"

	"github.com/0x7fffffff/verbatim/dashboard"
	"github.com/0x7fffffff/verbatim/megaphone"
	"github.com/0x7fffffff/verbatim/microphone"
	"github.com/0x7fffffff/verbatim/model"
)

// ActivityType The type of a given activity
type ActivityType int

// NotifyPortAdded lint
func NotifyPortAdded(portNum int, n model.Network) {
	log.Println("Port added", portNum, n.Name)
}

// Start I'm a stub.
func Start() {
	go dashboard.Start(nil)
	go microphone.Start(micListener{})
	go megaphone.Start(nil)

}

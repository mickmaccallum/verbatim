package relay

import (
	"fmt"
	"log"
	"net"

	"github.com/0x7fffffff/verbatim/model"
	// Lint
	_ "os"
	"sync"
)

// ActivityType The type of a given activity
type ActivityType int

const (
	// ActvityLogin lint
	ActvityLogin ActivityType = iota
	// ActivtySendData lint
	ActivtySendData
)

// Activity The type used for notifying the outside world of what is going on, at a high level.
type Activity struct {
	// ID lint
	ID   int
	Type ActivityType
}

var actvity chan Activity

// Current working assumption: Each captioner is going to be tied to all the encoders in a network
var networkToConns = make(map[model.Network]chan string)

var connMux = &sync.RWMutex{}

var encoderIDToConn = make(map[int]net.Conn)

// NotifyPortAdded lint
func NotifyPortAdded(portNum int, n model.Network) {
	log.Println("Port added", portNum, n.Name)
}

//
func notifyCaptionerSendingData() {

}

// TODO
func notifyBackendConnected() {

}

// TODO
func notifyBackendDisconnected() {

}

// TODO
func doBackendScheduledConnectAndDisconnect() {

}

// Start I'm a stub.
func Start() {

}

// AddDownstreamConnection Passing lint
func AddDownstreamConnection(encoder model.Encoder) error {
	// TODO: Dial into downstream connection
	addr := fmt.Sprint(encoder.IPAddress, ":", encoder.Port)
	log.Print(addr)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	// TODO: Move this
	conn.Close()
	defer connMux.Unlock()
	connMux.Lock()

	log.Panic("NOT DONE YET")
	return nil
}

/*

If we are going to be real time:

- When a network is added
-- How do we assign captioners to it?
--

- What captioners are online
-- For a given captioner, what network are they connected to?
-- How are we identifying captioners?
--

- Downstream state changes
-- Are we getting successful writes to the encoders?
-- How do we know if downstream is gone?
--
*/

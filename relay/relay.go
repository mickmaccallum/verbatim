package relay

import (
	"fmt"
	"log"
	"net"
	_ "os"
	"sync"
	"github.com/0x7fffffff/verbatim/persist"
)

// The type of a given activity
type ActivityType int

const (
	ActvityLogin ActivityType = iota
	ActivtySendData
)

// The type used for notifying the outside world of what is going on, at a high level.
type Activity struct {
	Id   int
	Type ActivityType
}

var actvity chan Activity

// Current working assumption: Each captioner is going to be tied to all the encoders in a network
var networkToConns = make(map[persist.Network]chan string)

var connMux = &sync.RWMutex{}

var encoderIdToConn = make(map[int]net.Conn)

func NotifyPortAdded(portNum int, n persist.Network) {
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

// Passing lint
func AddDownstreamConnection(encoder persist.Encoder) error {
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

package relay

import (
	"fmt"
	"github.com/0x7fffffff/verbatim/persist"
	"log"
	"net"
	_ "os"
	"sync"
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

// Need some kind of way to map between users and connections
// Need to use an RWMutex on this.
var encodersToConnections = make(map[persist.Encoder][]net.Conn)
var connMux = &sync.RWMutex{}

// TODO
// Connect this to a captioners list?
func notifyCaptionerConnected() {

}

// TODO
func notifyCaptionerSendingData() {

}

// TODO
func notifyBackendConnected() {

}

// TODO
func notifyBackenddisconnected() {

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

// Listen for TCP connections
func Start() error {
	ln, err := net.Listen("tcp", ":6000")
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleConnection(conn)
	}
}

// TODO: Get this working?
func authenticateUser(c net.Conn) error {
	return nil
}

func handleConnection(c net.Conn) {
	// TODO: Authenticate a new user
	authenticateUser(c)

	// TODO: Notify that a user is connected once authenticated
	buf := make([]byte, 4096)

	for {
		n, err := c.Read(buf)
		if err != nil || n == 0 {
			c.Close()
			break
		}
		n, err = c.Write(buf[0:n])
		if err != nil {
			c.Close()
			break
		}
	}
	log.Printf("Connection from %v closed.", c.RemoteAddr())
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

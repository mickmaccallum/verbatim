package relay

import (
	"fmt"
	"github.com/0x7fffffff/verbatim/persist"
	"log"
	"net"
	_ "os"
	"strconv"
	"strings"
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

// Current working assumption: Each captioner is going to be tied to all the encoders in a network
var networkToConns = make(map[persist.Network]chan string)

var connMux = &sync.RWMutex{}

var encoderIdToConn = make(map[int]net.Conn)

// TODO
// Connect this to a captioners list?
func notifyCaptionerConnected() {

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

// Listen for TCP connections
func Start() error {
	ln, err := net.Listen("tcp", "localhost:6000")
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

// This is now sort of working. We'll need to use this to work out how to send data around
func authenticateUser(c net.Conn) (*persist.Network, error) {
	buf := make([]byte, 128)
	_, err := c.Write([]byte("Enter the id of the network you want to connect to:"))
	if err != nil {
		c.Close()
		return nil, err
	}

	// Will read block?
	n, err := c.Read(buf)
	log.Println("Buff", string(buf[0:n]))
	if err != nil {
		c.Close()
	}

	// #runningwithscissors
	id, err := strconv.Atoi(strings.Trim(string(buf[0:n]), "\r\n\t "))
	if err != nil {
		fmt.Fprintln(c, "Invalid network id, please reconnect and try again")
		log.Println(err.Error())
		c.Close()
		return nil, err
	}

	network, err := persist.GetNetwork(id)
	if err != nil {
		fmt.Fprintln(c, "Unable to find network!")
		c.Close()
		return nil, err
	}
	return network, nil
}

// Basic demoable state.
func handleConnection(c net.Conn) {
	// TODO: Authenticate a new user
	network, err := authenticateUser(c)
	if err != nil {
		// Bail
		return
	}

	fmt.Fprintln(c, "Connected to echo server for:", network.Name)

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

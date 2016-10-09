package microphone

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/0x7fffffff/verbatim/persist"
)

// This information tells us
type CaptionerInfo struct {
	numConn int
	IPAddr  net.IP
}

type CaptionerStatus struct {
	CaptionerInfo
	lastActivity time.Time
}

// These are the events that the server using this will be notified of
type CaptionListener interface {
	// When a captioner connects
	Connected(CaptionerInfo, persist.Network)
	Disconnected(CaptionerInfo, persist.Network)
	// Send a data message to the network
	SendDataToNetwork(persist.Network, []byte)
}

// This is our reference to the delegate methods from the relay server
var l CaptionListener

var networks map[persist.Network][]CaptionerStatus

func GetListnerStatus() map[persist.Network]CaptionerStatus {
	return nil
}

// Listen for TCP connections
func Start(listener CaptionListener) error {
	l = listener
	networks, err := persist.GetNetworks()
	if err != nil {
		return err
	}
	for _, n := range networks {
		go listenForNetwork(n)
	}
	return nil
}

// TODO: Error handling here?
func AddNetwork(n persist.Network) {
	// Hrm...
	go listenForNetwork(n)
}

var connsPerIP = make(map[string]int)
var connsPerIPMux = &sync.RWMutex{}

// TODO A way to report failed listers?
func listenForNetwork(n persist.Network) {
	var connsPerIP = make(map[string]int)
	ln, err := net.Listen("tcp", fmt.Sprint(":", n.Port))
	if err != nil {
		// What to do, try to reconnect with exponential backoff, or
		// let the user decide what to do?
		log.Fatal(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		connsPerIP[conn.RemoteAddr().String()]++
		go handleCaptioner(conn, n)
	}
}

// Basic demoable state.
func handleCaptioner(c net.Conn, network persist.Network) {
	// Keep a buffer of 1KiB per captioner
	buf := make([]byte, 1024)
	for {
		n, err := c.Read(buf)
		if err != nil || n == 0 {
			c.Close()
			break
		}
		// Send any recieved bytes to the relay server
		l.SendDataToNetwork(network, buf[0:n])
	}
	log.Printf("Connection from %v closed.", c.RemoteAddr())
}

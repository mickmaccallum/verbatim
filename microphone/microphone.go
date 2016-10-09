package microphone

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/0x7fffffff/verbatim/model"
	"github.com/0x7fffffff/verbatim/persist"
)

// This information tells us
type CaptionerInfo struct {
	numConn int
	IPAddr  string
	network model.Network
}

type CaptionerStatus struct {
	CaptionerInfo
	lastActivity time.Time
}

// These are the events that the server using this will be notified of
type CaptionListener interface {
	// When a captioner connects
	Connected(CaptionerInfo, model.Network)
	Disconnected(CaptionerInfo, model.Network)
	// Send a data message to the network
	SendDataToNetwork(model.Network, []byte)
}

// This is our reference to the delegate methods from the relay server
var l CaptionListener

func GetListnerStatus() map[model.Network]map[CaptionerInfo]struct{} {
	askStatus <- struct{}{}
	return <-getStatus
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
func AddNetwork(n model.Network) {
	// Hrm...
	go listenForNetwork(n)
}

var addNetworkChan = make(chan model.Network, 10)
var addListenerChan = make(chan CaptionerInfo, 10)
var rmNetworkChan = make(chan model.Network, 10)
var rmListenerChan = make(chan CaptionerInfo, 10)

var askStatus = make(chan struct{})
var getStatus = make(chan map[model.Network]map[CaptionerInfo]struct{})

// This function is the sole arbiter of state for these stats
func maintainListenerState() {
	networks := make(map[model.Network]map[CaptionerInfo]struct{})
	for {
		select {
		case n := <-addNetworkChan:
			networks[n] = make(map[CaptionerInfo]struct{})
		case l := <-addListenerChan:
			(networks[l.network])[l] = struct{}{}
		case n := <-rmNetworkChan:
			delete(networks, n)
		case l := <-rmListenerChan:
			delete(networks[l.network], l)
		case <-askStatus:
			cp := make(map[model.Network]map[CaptionerInfo]struct{})
			for key, value := range networks {
				cp[key] = make(map[CaptionerInfo]struct{})
				for ik, iv := range value {
					cp[key][ik] = iv
				}
			}
			getStatus <- cp
		}
	}
}

var connsPerIP = make(map[string]int)
var connsPerIPMux = &sync.RWMutex{}

// TODO A way to report failed listers?
func listenForNetwork(n model.Network) {
	var connsPerIP = make(map[string]int)
	ln, err := net.Listen("tcp", fmt.Sprint(":", n.ListeningPort))
	if err != nil {
		// What to do, try to reconnect with exponential backoff, or
		// let the user decide what to do?
		log.Fatal(err)
	}
	addNetworkChan <- n
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		key := conn.RemoteAddr().String()
		if val, found := connsPerIP[key]; found {
			connsPerIP[key] = val + 1
		} else {
			connsPerIP[key] = 1
		}
		go handleCaptioner(conn, n)
	}
}

// Basic demoable state.
func handleCaptioner(c net.Conn, network model.Network) {

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

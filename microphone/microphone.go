package microphone

import (
	"fmt"
	"github.com/0x7fffffff/verbatim/megaphone"
	"log"
	"math"
	"net"
	"time"

	"github.com/0x7fffffff/verbatim/model"
	"github.com/0x7fffffff/verbatim/persist"
)

type CaptionerID struct {
	IPAddr  string
	NumConn int
}

func (c CaptionerID) String() string {
	return fmt.Sprint(c.IPAddr, ":", c.NumConn)
}

// The connection information for a given captioner.
type CaptionerInfo struct {
	CaptionerID
	network model.Network
}

type CaptionerStatus struct {
	CaptionerInfo
	lastActivity time.Time
}

// These are the events that the server using this will be notified of
type CaptionListener interface {
	// When a captioner connects
	Connected(ci CaptionerInfo)
	// Report that a captioner has been disconnected
	Disconnected(ci CaptionerInfo)
	// Report that a captioner has been successfully muted
	Muted(ci CaptionerInfo)

	// Ask the relay server for a network writer
	GetBroadcaster(network model.Network) megaphone.NetworkBroadcaster
}

// This is our reference to the delegate methods from the relay server
var relay CaptionListener

// Get the listener status for the entire contigent of listeners
func GetListnerStatus() map[model.Network]map[CaptionerInfo]struct{} {
	askStatus <- struct{}{}
	return <-getStatus
}

// Listen for TCP connections
func Start(listener CaptionListener) error {
	relay = listener
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

func MuteCaptioner(n model.Network, captionerId int) {

}

// Paired channels
var (
	addNetworkChan   = make(chan model.Network, 10)
	rmNetworkChan    = make(chan model.Network, 10)
	addListenerChan  = make(chan CaptionerInfo, 10)
	rmListenerChan   = make(chan CaptionerInfo, 10)
	muteListenChan   = make(chan CaptionerInfo, 10)
	unmuteListenChan = make(chan CaptionerInfo, 10)
)

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
			network := networks[l.network]
			network[l] = struct{}{}
			relay.Connected(l, l.network)
		case n := <-rmNetworkChan:
			delete(networks, n)
		case l := <-rmListenerChan:
			delete(networks[l.network], l)
			relay.Disconnected(l, l.network)
		case <-askStatus:
			// Copy all the current status and send it off to the relay server,
			// to avoid coherency isues
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

// TODO A way to report failed listers?
func listenForNetwork(n model.Network) {
	var connsPerIP = make(map[string]int)
	ln, err := net.Listen("tcp", fmt.Sprint(":", n.ListeningPort))
	if err != nil {
		// What to do, try to reconnect with exponential backoff, or
		// let the user decide what to do?
		log.Fatal(err)
	}
	// Add the network to the list of networks in use.
	addNetworkChan <- n
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		key := conn.RemoteAddr().String()
		if val, found := connsPerIP[key]; found {
			connsPerIP[key] = (val + 1) % math.MaxInt32
		} else {
			connsPerIP[key] = 1
		}
		val := connsPerIP[key]
		go handleCaptioner(conn, n, val)
	}
}

// Basic demoable state.
func handleCaptioner(c net.Conn, network model.Network, numConn int) {
	// Notify that we have a new listener
	capInfo := CaptionerInfo{
		numConn: numConn,
		IPAddr:  c.RemoteAddr().String(),
		network: network,
	}
	addListenerChan <- capInfo
	// Keep a buffer of 1KiB per captioner
	buf := make([]byte, 1024)
	for {
		c.SetReadDeadline(time.Now().Add(time.Minute * 10))
		n, err := c.Read(buf)
		if err != nil || n == 0 {
			c.Close()
			rmListenerChan <- capInfo
			log.Println(err.Error())
			break
		}
		// Copy the data to make sure the
		// byte slice doesn't get changed under the encoder sending
		// it out.
		message := make([]byte, n)
		copy(message, bug[0:n])
		// Send any recieved bytes to the relay server
		relay.SendDataToNetwork(network, message)
	}
	log.Printf("Connection from %v closed.", c.RemoteAddr())
}

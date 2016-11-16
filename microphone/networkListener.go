package microphone

import (
	"fmt"
	"github.com/0x7fffffff/verbatim/model"
	"log"
	"math"
	"net"
	"time"
)

type networkListeningServer struct {

	// Indicates if this server is active, or shutting down
	isActive bool
	// The listener for this network
	ln net.Listener
	// Pull in the id and such from here.
	network model.Network

	// All the currently connected captioners
	captioners map[model.CaptionerID]CaptionListener

	// If we get a change request, we need to know what to change to.
	tryPortChange chan int
	// Indicate if the change could successfully be made.
	couldMakePortChange chan error

	// Change the timeout for a listen call...
	changeTimeout     chan time.Time
	timeoutChanged    chan error
	tryAddCaptioner   chan CaptionListener
	couldAddCaptioner chan error
	rmCaptioner       chan model.CaptionerID
	muteCaptioner     chan model.CaptionerID
	unmuteCaptioner   chan model.CaptionerID
	killSelf          chan struct{}
}

// Attempt to spin up a listener, and return if it succeeded
func tryMakeNetworkListener(n model.Network) (*networkListeningServer, error) {
	var ln net.Listener
	var err error
	if ln, err = attemptListen(n); err != nil {
		return nil, err
	}

	return &networkListeningServer{
		isActive:            true,
		ln:                  ln,
		captioners:          make(map[model.CaptionerID]CaptionListener),
		tryPortChange:       make(chan int),
		couldMakePortChange: make(chan error),
		changeTimeout:       make(chan time.Time),
		timeoutChanged:      make(chan error),
		tryAddCaptioner:     make(chan CaptionListener),
		couldAddCaptioner:   make(chan error),
		rmCaptioner:         make(chan model.CaptionerID),
		muteCaptioner:       make(chan model.CaptionerID),
		unmuteCaptioner:     make(chan model.CaptionerID),
		killSelf:            make(chan struct{}),
	}, nil
}

func (n networkListeningServer) MuteCaptioner(id model.CaptionerID) {
	n.muteCaptioner <- id
}

func (n networkListeningServer) UnmuteCaptioner(id model.CaptionerID) {
	n.unmuteCaptioner <- id
}

// Maintain all the state related to a network
func (n *networkListeningServer) serve() {
	for {
		select {
		case <-n.tryPortChange:
		case <-n.changeTimeout:
		case <-n.tryAddCaptioner:
			// TODO: Port this code to work in this server loop
			/*
				n.captioners[model.CaptionerID(cl.cell.id)] = cl
				if arr, found := listenersByNetwork[model.NetworkID(cl.NetId)]; found {
					arr = append(arr, cl)
					if len(arr) == 1 {
						cl.cell.Unmute()
					}
					listenersByNetwork[cl.NetId] = arr
					relay.Connected(cl.cell.id)
					cl.cell.cellMux.Lock()
					if cl.cell.isMute {
						relay.Muted(cl.cell.id)
					} else {
						relay.Unmuted(cl.cell.id)
					}
					cl.cell.cellMux.Unlock()
					couldAddCaptioner <- nil
				} else {
					couldAddCaptioner <- fmt.Errorf("")
				}
			*/
		case <-n.rmCaptioner:
		case muteID := <-n.muteCaptioner:
			if cl, found := n.captioners[muteID]; found {
				cl.cell.Mute()
				relay.Muted(muteID)
			}
		case <-n.unmuteCaptioner:

		}
	}
}

func (n networkListeningServer) Close() {
	n.killSelf <- struct{}{}
}

func attemptListen(n model.Network) (net.Listener, error) {
	return net.Listen("tcp", fmt.Sprint(":", n.ListeningPort))
}

func (srv networkListeningServer) handleCaptioner(c net.Conn, writer *MuteCell) {
	// Notify that we have a new listener
	// addListenerChan <-
	// Keep a buffer of 1KiB per captioner
	buf := make([]byte, 1024)
	log.Println("Am listening to captioner")
	for {
		c.SetReadDeadline(time.Now().Add(time.Second * 30))
		n, err := c.Read(buf)
		if err != nil || n == 0 {
			log.Println("Disconnected from Captioner")
			srv.rmCaptioner <- writer.id
			log.Println(err.Error())
			break
		}
		// Copy the data to make sure the
		// byte slice doesn't get changed under the encoder sending
		// it out.
		message := make([]byte, n)
		copy(message, buf[0:n])
		writer.Write(message)
		// Send any recieved bytes to the relay server
	}
	// log.Printf("Connection from %v closed.", c.RemoteAddr())
}

// TODO A way to report failed listers?
func (srv networkListeningServer) listenForNetwork(n model.Network, ln net.Listener) {
	var connsPerIP = make(map[string]int)
	// Add the network to the list of networks in use.
	broadcaster := relay.GetBroadcaster(n)
	for {
		conn, er := ln.Accept()
		if er != nil {
			log.Println("Connection failed:", er.Error())
			if err := er.(net.Error); err != nil {
				if err.Temporary() {
					log.Println(err)
					continue
				} else {
					// TODO: Signal error here.
					relay.NetworkListenFailed(n)
					return
				}
			}
		}

		// Make sure that a given captioner has a way of being identified
		key := conn.RemoteAddr().String()
		if val, found := connsPerIP[key]; found {
			connsPerIP[key] = (val + 1) % math.MaxInt32
		} else {
			connsPerIP[key] = 1
		}
		val := connsPerIP[key]
		writer := makeMuteCell(broadcaster, model.CaptionerID{
			NumConn:   val,
			IPAddr:    conn.RemoteAddr().String(),
			NetworkID: model.NetworkID(n.ID),
		})
		srv.tryAddCaptioner <- CaptionListener{
			conn:  conn,
			cell:  writer,
			NetId: n.ID,
		}
		if <-srv.couldAddCaptioner != nil {
			return
		}
		go srv.handleCaptioner(conn, writer)
	}
}

package microphone

import (
	"fmt"
	"github.com/0x7fffffff/verbatim/model"
	"github.com/0x7fffffff/verbatim/states"
	"log"
	"math"
	"net"
	"time"
)

type CaptionListener struct {
	NetId model.NetworkID
	port  int
	conn  net.Conn
	cell  *MuteCell
}

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
	changeTimeout     chan time.Duration
	timeoutChanged    chan error
	tryAddCaptioner   chan CaptionListener
	couldAddCaptioner chan error
	rmCaptioner       chan model.CaptionerID
	muteCaptioner     chan model.CaptionerID
	unmuteCaptioner   chan model.CaptionerID
	killSelf          chan struct{}
	// These are paired channels
	askCaptioners chan struct{}
	captionerList chan []CaptionerStatus
}

// Attempt to spin up a listener, and return if it succeeded
func tryMakeNetworkListener(n model.Network) (*networkListeningServer, error) {
	var ln net.Listener
	var err error
	if ln, err = attemptListen(n.ListeningPort); err != nil {
		return nil, err
	}

	return &networkListeningServer{
		isActive:            true,
		ln:                  ln,
		network:             n,
		captioners:          make(map[model.CaptionerID]CaptionListener),
		tryPortChange:       make(chan int),
		couldMakePortChange: make(chan error),
		changeTimeout:       make(chan time.Duration),
		timeoutChanged:      make(chan error),
		tryAddCaptioner:     make(chan CaptionListener),
		couldAddCaptioner:   make(chan error),
		rmCaptioner:         make(chan model.CaptionerID),
		muteCaptioner:       make(chan model.CaptionerID),
		unmuteCaptioner:     make(chan model.CaptionerID),
		killSelf:            make(chan struct{}),
		askCaptioners:       make(chan struct{}),
		captionerList:       make(chan []CaptionerStatus),
	}, nil
}

func (n networkListeningServer) MuteCaptioner(id model.CaptionerID) {
	n.muteCaptioner <- id
}

func (n networkListeningServer) UnmuteCaptioner(id model.CaptionerID) {
	n.unmuteCaptioner <- id
}

func (n networkListeningServer) TryChangePort(port int) error {
	n.tryPortChange <- port
	return <-n.couldMakePortChange
}

func (n networkListeningServer) RemoveCaptioner(id model.CaptionerID) {
	n.rmCaptioner <- id
}

func (n networkListeningServer) ChangeTimeout(seconds int) {
	n.changeTimeout <- (time.Duration(seconds) * time.Second)
}

func (n networkListeningServer) GetConnectedCaptioners() []CaptionerStatus {
	n.askCaptioners <- struct{}{}
	return <-n.captionerList
}

func (n networkListeningServer) Close() {
	n.killSelf <- struct{}{}
}

var (
	errServerStopped               = fmt.Errorf("This server has been closed, it is not accpeting new connections")
	errCaptionersAreStillConnected = fmt.Errorf("Captioners are still connected to this network server")
)

// Maintain all the state related to a network
func (n *networkListeningServer) serve() {
	// Kick off the listening loop
	go n.listenForNetwork()
	for {
		select {
		case <-n.killSelf:
			n.isActive = false
			for _, cl := range n.captioners {
				cl.cell.Mute()
				cl.conn.Close()
			}
			n.ln.Close()
			n.ln = nil
		case port := <-n.tryPortChange:
			if len(n.captioners) > 0 {
				n.couldMakePortChange <- errCaptionersAreStillConnected
			} else {
				if ln, err := attemptListen(port); err != nil {
					n.couldMakePortChange <- err
				} else {
					// Closing the old connection should cause a loop exit
					n.ln.Close()
					n.ln = ln
					n.network.ListeningPort = port
					n.couldMakePortChange <- nil
					go n.listenForNetwork()
					relay.NetworkListenSucceeded(n.network)
				}
			}
		case timeoutLen := <-n.changeTimeout:
			for _, cl := range n.captioners {
				// Reset each timeout to be in the future
				cl.conn.SetDeadline(cl.cell.LastWaitTime().Add(time.Second * timeoutLen))
			}
		case cl := <-n.tryAddCaptioner:
			if n.isActive && cl.port == n.network.ListeningPort {
				// Mute any existing captioners
				for _, cl := range n.captioners {
					cl.cell.Mute()
				}
				cl.cell.Unmute()
				n.couldAddCaptioner <- nil
				n.captioners[cl.cell.id] = cl
				relay.Connected(cl.cell.id)
				relay.Unmuted(cl.cell.id)
			} else {
				n.couldAddCaptioner <- errServerStopped
			}
		case <-n.askCaptioners:
			captionersToReturn := make([]CaptionerStatus, len(n.captioners))
			i := 0
			for id, cl := range n.captioners {
				captionersToReturn[i].ID = id
				cl.cell.cellMux.Lock()
				if cl.cell.isMute {
					captionersToReturn[i].State = states.CaptionerMuted
				} else {
					captionersToReturn[i].State = states.CaptionerUnmuted
				}
				cl.cell.cellMux.Unlock()
				i++
			}
			n.captionerList <- captionersToReturn
		case rmId := <-n.rmCaptioner:
			if cl, found := n.captioners[rmId]; found {
				cl.cell.Mute()
				cl.conn.Close()
				delete(n.captioners, rmId)
			}
		case muteID := <-n.muteCaptioner:
			if cl, found := n.captioners[muteID]; found {
				cl.cell.Mute()
				relay.Muted(muteID)
			}
		case unmuteID := <-n.unmuteCaptioner:
			if cl, found := n.captioners[unmuteID]; found {
				cl.cell.Unmute()
				relay.Unmuted(unmuteID)
			}
		}
	}
}

func attemptListen(port int) (net.Listener, error) {
	return net.Listen("tcp", fmt.Sprint(":", port))
}

func (srv networkListeningServer) handleCaptioner(c net.Conn, writer *MuteCell) {
	// Notify that we have a new listener
	// Keep a buffer of 1KiB per captioner
	buf := make([]byte, 1024)
	// log.Println("Am listening to captioner")
	for {
		now := time.Now()
		// log.Println(srv.network)
		c.SetReadDeadline(now.Add(time.Duration(srv.network.Timeout) * time.Second))
		writer.SetWaitTime(now)
		n, err := c.Read(buf)
		if err != nil || n == 0 {
			log.Println("Disconnected from Captioner")
			srv.rmCaptioner <- writer.id
			log.Println(err.Error())
			break
		}
		// Copy the data to make sure the
		// byte slice doesn't get changed under the encoder sending it out.
		message := make([]byte, n)
		copy(message, buf[0:n])
		writer.Write(message)
		// Send any recieved bytes to the relay server
	}
}

// TODO A way to report failed listers?
func (srv networkListeningServer) listenForNetwork() {
	var connsPerIP = make(map[string]int)
	// Add the network to the list of networks in use.
	broadcaster := relay.GetBroadcaster(srv.network)
	for {
		conn, er := srv.ln.Accept()
		if er != nil {
			log.Println("Connection failed:", er.Error())
			if err := er.(net.Error); err != nil {
				if err.Temporary() {
					log.Println(err)
					continue
				} else {
					// TODO: Signal error here.
					relay.NetworkListenFailed(srv.network)
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
			NetworkID: model.NetworkID(srv.network.ID),
		})
		srv.tryAddCaptioner <- CaptionListener{
			conn:  conn,
			port:  srv.network.ListeningPort,
			cell:  writer,
			NetId: srv.network.ID,
		}
		// Don't spin up a caption listener
		if <-srv.couldAddCaptioner != nil {
			return
		}
		go srv.handleCaptioner(conn, writer)
	}
}

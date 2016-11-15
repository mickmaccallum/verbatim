package microphone

import (
	"github.com/0x7fffffff/verbatim/model"
	"net"
)

type networkListeningServer struct {

	// Indicates if this server is active, or shutting down
	bool isActive
	// The listener for this network
	ln net.Listener
	// Pull in the id and such from here.
	network         model.Network
	captionerAdded  chan CaptionListener
	rmCaptioner     chan model.CaptionerID
	muteCaptioenr   chan model.CaptionerID
	unmuteCaptioner chan model.CaptionerID
}

// Maintain all the state related to a network
func (n *networkListeningServer) serve() {
	for {
		select {}
	}
}

func attemptListen(n model.Network) (net.Listener, error) {
	return net.Listen("tcp", fmt.Sprint(":", n.ListeningPort))
}

func handleCaptioner(c net.Conn, writer *MuteCell) {
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
			rmCaptioner <- writer.id
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
func listenForNetwork(n model.Network, ln net.Listener) {
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
					relay.NetworkListenFailed(network)
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
		captionerAdded <- CaptionListener{
			conn:  conn,
			cell:  writer,
			NetId: n.ID,
		}
		go handleCaptioner(conn, writer)
	}
}

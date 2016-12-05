package megaphone

import (
	"context"
	"fmt"
	"github.com/0x7fffffff/verbatim/model"
	"github.com/0x7fffffff/verbatim/persist"
	"log"
	"net"
	"time"
)

type AddEncoderResult int

const (
	encoderDidExist AddEncoderResult = iota
	encoderDidNotExist
)

type encoderConn struct {
	channel chan []byte
	cancel  context.CancelFunc
}

type encoderChan struct {
	id model.EncoderID
	encoderConn
}

type encoderIdPair struct {
	network model.NetworkID
	encoder model.EncoderID
}

type NetworkBroadcaster struct {
	id             model.NetworkID
	writeChan      chan []byte
	encoderExisted chan AddEncoderResult
	addEncoder     chan model.Encoder
	rmEncoder      chan model.EncoderID
	// faultedEncoder chan model.EncoderID
	// restartEncoder chan encoderIdPair
	die         chan struct{}
	getEncoders chan struct{}
	encoderIds  chan []model.EncoderID
	encoders    map[model.EncoderID]encoderConn
}

func makeBroadcaster(n model.NetworkID /*, restartEncoder chan encoderIdPair*/) *NetworkBroadcaster {
	return &NetworkBroadcaster{
		id:             n,
		writeChan:      make(chan []byte, 10),
		encoderExisted: make(chan AddEncoderResult),
		addEncoder:     make(chan model.Encoder),
		rmEncoder:      make(chan model.EncoderID),
		encoders:       make(map[model.EncoderID]encoderConn),
		die:            make(chan struct{}),
		// faultedEncoder: make(chan model.EncoderID),
		// restartEncoder: restartEncoder,
		getEncoders: make(chan struct{}),
		encoderIds:  make(chan []model.EncoderID),
	}
}

// So do I defer to catch a panic here if the
func (n NetworkBroadcaster) Write(buf []byte) {
	n.writeChan <- buf
}

func (n NetworkBroadcaster) removeEncoder(id model.EncoderID) {
	n.rmEncoder <- id
}

// If it returns true, the encoder was added, otherwise it was already running
func (n NetworkBroadcaster) registerEncoder(enc model.Encoder) AddEncoderResult {
	n.addEncoder <- enc
	return <-n.encoderExisted
}

func (n NetworkBroadcaster) getConnectedEncoderIds() []model.EncoderID {
	n.getEncoders <- struct{}{}
	return <-n.encoderIds
}

func (n NetworkBroadcaster) destroy() {
	n.die <- struct{}{}
}

// TODO: Figure out a way to panic if launched more than once.
// should be launched in a goroutine
func (n *NetworkBroadcaster) serveConnection() {
	for {
		select {
		case buf := <-n.writeChan:
			if len(n.encoders) == 0 {
				// Log here
				persist.CreateBackup(buf, n.id)
			} else {
				for _, dest := range n.encoders {
					dest.channel <- buf
				}
			}
		case enc := <-n.addEncoder:
			// Only add an encoder if it hasn't been added already
			if _, found := n.encoders[enc.ID]; found {
				n.encoderExisted <- encoderDidExist
				continue
			} else {
				dest := make(chan []byte)
				ctx, cancel := context.WithCancel(context.Background())
				n.encoders[enc.ID] = encoderConn{dest, cancel}
				go handleEncoder(enc, dest, ctx, n)
				n.encoderExisted <- encoderDidNotExist
			}
			// backoffs[dest.id] = 1 * time.Microsecond
		case id := <-n.rmEncoder:
			if val, found := n.encoders[id]; found {
				val.cancel()
				delete(n.encoders, id)
			}
		case <-n.getEncoders:
			encoders := make([]model.EncoderID, 0)
			for id := range n.encoders {
				encoders = append(encoders, id)
			}
			n.encoderIds <- encoders
		case <-n.die:
			// Close all the encoders hooked up to this broadcaster
			close(n.writeChan)
			for _, ch := range n.encoders {
				// Cancel out all the connections
				ch.cancel()
			}
			return
		}
	}
}

const LINE_CUT_WIDTH = 32

func writeMessageSegmented(conn net.Conn, msg []byte) error {
	// Write message in chunks
	for i := 0; i*LINE_CUT_WIDTH < len(msg); i++ {
		begin := i * LINE_CUT_WIDTH
		end := (i + 1) * LINE_CUT_WIDTH
		if end > len(msg) {
			end = len(msg)
		}
		if _, err := conn.Write(msg[begin:end]); err != nil {
			return err
		}
	}
	return nil
}

func loginToEncoder(enc model.Encoder, ctx context.Context) (net.Conn, error) {
	addr := fmt.Sprint(enc.IPAddress, ":", enc.Port)
	// Give a 5 second timeout for the dial to succeed or fail
	conn, err := (&net.Dialer{
		Timeout: time.Second * 5,
	}).DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}
	// Hacky, make sure this isn't broken?
	go func() {
		<-(ctx.Done())
		conn.Close()
	}()
	if _, err = conn.Write([]byte(enc.Handle + "\n")); err != nil {
		return nil, err
	}
	if _, err = conn.Write([]byte(enc.Password + "\n")); err != nil {
		return nil, err
	}
	return conn, nil
}

func handleEncoder(enc model.Encoder, inbound chan []byte, ctx context.Context, n *NetworkBroadcaster) {
	relay.LoggingIn(enc)
	conn, err := loginToEncoder(enc, ctx)
	if err != nil {
		// Login failed, remove it from the list of the things
		n.removeEncoder(encId(enc))
		// And then notify that login failed for the encoder
		// Allowing the user to try to relogin
		relay.LoginFailed(enc)
		// conn.Close()
		return
	}
	relay.LoginSucceeded(enc)
	for {
		select {
		case <-(ctx.Done()):
			log.Println("Encoder removed")
			close(inbound)
			relay.Logout(enc)
			return

		case msg, ok := <-inbound:
			if ok {
				err := writeMessageSegmented(conn, msg)
				if err != nil {
					// Signal to the broadcaster that we have an error
					relay.UnexpectedDisconnect(enc)
					n.removeEncoder(enc.ID)
					conn.Close()
					// n.faultedEncoder <- encId(enc)
					return
				}
			}
		}
	}
}

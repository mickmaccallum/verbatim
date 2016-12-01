package megaphone

import (
	"github.com/0x7fffffff/verbatim/model"
	"github.com/0x7fffffff/verbatim/persist"
)

type AddEncoderResult int

const (
	encoderDidExist AddEncoderResult = iota
	encoderDidNotExist
)

type encoderChan struct {
	id      model.EncoderID
	channel chan []byte
}

type encoderIdPair struct {
	network model.NetworkID
	encoder model.EncoderID
}

type NetworkBroadcaster struct {
	id             model.NetworkID
	writeChan      chan []byte
	encoderExisted chan bool
	addEncoder     chan encoderChan
	rmEncoder      chan model.EncoderID
	faultedEncoder chan model.EncoderID
	restartEncoder chan encoderIdPair
	die            chan struct{}
	getEncoders    chan struct{}
	encoderIds     chan []model.EncoderID
	encoders       map[model.EncoderID]chan []byte
}

func makeBroadcaster(n model.NetworkID, restartEncoder chan encoderIdPair) *NetworkBroadcaster {
	return &NetworkBroadcaster{
		id:             n,
		writeChan:      make(chan []byte, 10),
		encoderExisted: make(chan bool),
		addEncoder:     make(chan encoderChan),
		rmEncoder:      make(chan model.EncoderID),
		encoders:       make(map[model.EncoderID]chan []byte),
		die:            make(chan struct{}),
		faultedEncoder: make(chan model.EncoderID),
		restartEncoder: restartEncoder,
		getEncoders:    make(chan struct{}),
		encoderIds:     make(chan []model.EncoderID),
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
func (n NetworkBroadcaster) registerEncoderChan(id model.EncoderID, dest chan []byte) AddEncoderResult {
	n.addEncoder <- encoderChan{id, dest}
	if <-n.encoderExisted {
		return encoderDidExist
	} else {
		return encoderDidNotExist
	}
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
					dest <- buf
				}
			}
		case dest := <-n.addEncoder:
			// Only add an encoder if it hasn't been added already
			if _, found := n.encoders[dest.id]; found {
				n.encoderExisted <- true
				continue
			} else {
				n.encoders[dest.id] = dest.channel
				n.encoderExisted <- false
			}
			// backoffs[dest.id] = 1 * time.Microsecond
		case id := <-n.faultedEncoder:
			close(n.encoders[id])
			delete(n.encoders, id)
		case id := <-n.rmEncoder:
			if _, found := n.encoders[id]; found {
				close(n.encoders[id])
				delete(n.encoders, id)
			}
		case <-n.getEncoders:
			encoders := make([]model.EncoderID, 0)
			for id := range n.encoders {
				encoders = append(encoders, id)
			}
			n.encoderIds <- encoders
		case <-n.die:
			// Close all the encoders hooked up to this broa
			close(n.writeChan)
			for _, ch := range n.encoders {
				close(ch)
			}
			return
		}
	}
}

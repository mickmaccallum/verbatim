package megaphone

import (
	"time"
)

type encoderChan struct {
	id      EncoderID
	channel chan []byte
}

type encoderIdPair struct {
	network NetworkID
	encoder EncoderID
}

type NetworkBroadcaster struct {
	id             NetworkID
	writeChan      chan []byte
	addEncoder     chan encoderChan
	rmEncoder      chan EncoderID
	faultedEncoder chan EncoderID
	restartEncoder chan encoderIdPair
	die            chan struct{}
	encoders       map[EncoderID]chan []byte
}

func makeBroadcaster(n NetworkID, restartEncoder chan encoderIdPair) *NetworkBroadcaster {
	return &NetworkBroadcaster{
		id:             n,
		writeChan:      make(chan []byte, 10),
		addEncoder:     make(chan encoderChan),
		rmEncoder:      make(chan EncoderID),
		encoders:       make(map[EncoderID]chan []byte),
		die:            make(chan struct{}),
		faultedEncoder: make(chan EncoderID),
		restartEncoder: restartEncoder,
	}
}

// So do I defer to catch a panic here if the
func (n NetworkBroadcaster) Write(buf []byte) {
	n.writeChan <- buf
}

func (n NetworkBroadcaster) removeEncoder(id EncoderID) {
	n.rmEncoder <- id
}

func (n NetworkBroadcaster) registerEncoderChan(id EncoderID, dest chan []byte) {
	n.addEncoder <- encoderChan{id, dest}
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
			for _, dest := range n.encoders {
				dest <- buf
			}
		case dest := <-n.addEncoder:
			n.encoders[dest.id] = dest.channel
			// backoffs[dest.id] = 1 * time.Microsecond
		case id := <-n.faultedEncoder:
			close(n.encoders[id])
			delete(n.encoders, id)
			// WARNING:
			go func() {
				// TODO: Implement expoential backoff here.
				<-time.After(5 * time.Second)
				n.restartEncoder <- encoderIdPair{n.id, id}
			}()
		case id := <-n.rmEncoder:
			close(n.encoders[id])
			delete(n.encoders, id)
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

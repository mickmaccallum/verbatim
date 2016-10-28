package microphone

import (
	"github.com/0x7fffffff/verbatim/megaphone"
	"github.com/0x7fffffff/verbatim/model"
	"sync"
)

type MuteCell struct {
	id          model.CaptionerID
	isMute      bool
	cellMux     *sync.Mutex
	broadcaster *megaphone.NetworkBroadcaster
}

func makeMuteCell(b *megaphone.NetworkBroadcaster, id model.CaptionerID) *MuteCell {
	return &MuteCell{
		isMute:      false,
		cellMux:     &sync.Mutex{},
		broadcaster: b,
		id:          id,
	}
}

func (c *MuteCell) Mute() {
	c.cellMux.Lock()
	c.isMute = true
	c.cellMux.Unlock()
}

func (c *MuteCell) Unmute() {
	c.cellMux.Lock()
	c.isMute = false
	c.cellMux.Unlock()
}

func (c *MuteCell) Write(buf []byte) {
	c.cellMux.Lock()
	if !c.isMute {
		c.broadcaster.Write(buf)
	}
	c.cellMux.Unlock()
}

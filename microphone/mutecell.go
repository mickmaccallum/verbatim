package microphone

import (
	"github.com/0x7fffffff/verbatim/megaphone"
	"github.com/0x7fffffff/verbatim/model"
	"sync"
	"time"
)

type MuteCell struct {
	id           model.CaptionerID
	isMute       bool
	cellMux      *sync.Mutex
	broadcaster  *megaphone.NetworkBroadcaster
	lastWaitTime time.Time
}

func makeMuteCell(b *megaphone.NetworkBroadcaster, id model.CaptionerID) *MuteCell {
	return &MuteCell{
		isMute:      true,
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

func (c *MuteCell) SetWaitTime(t time.Time) {
	c.cellMux.Lock()
	c.lastWaitTime = t
	c.cellMux.Unlock()
}

func (c *MuteCell) LastWaitTime() time.Time {
	c.cellMux.Lock()
	retval := c.lastWaitTime
	c.cellMux.Unlock()
	return retval
}

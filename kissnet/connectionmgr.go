package kissnet

import (
	"sync"
	"sync/atomic"
)

type ConnectionMgr struct {
	sesMgr   sync.Map
	sesIDGen int64
	count    int64
}

func (this *ConnectionMgr) add(c IConnection) {
	atomic.AddInt64(&this.count, 1)
	id := atomic.AddInt64(&this.sesIDGen, 1)
	this.sesMgr.Store(id, c)
}

func (this *ConnectionMgr) getConnection(id int64) IConnection {
	if v, ok := this.sesMgr.Load(id); ok {
		return v.(IConnection)
	}
	return nil
}
func (this *ConnectionMgr) del(c IConnection) {
	this.sesMgr.Delete(c.getID())
	atomic.AddInt64(&this.count, -1)
}

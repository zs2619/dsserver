package gameplayer

import (
	"dsservices/kissnet"
	"dsservices/pb"
	"fmt"
	"sync"
)

var GPlayerMgr *PlayerMgr = &PlayerMgr{
	playerIDMgr:       make(map[string]*GamePlayer),
	playerClientIDMgr: make(map[kissnet.IConnection]*GamePlayer),
	num:               int64(0),
}

func ProcMsg(conn kissnet.IConnection, msgID int32, msg []byte) error {
	f, ok := msgMap[pb.C2S_MsgID_MsgID(msgID)]
	if !ok {
		return fmt.Errorf("(%d) msgID nil", msgID)
	}
	p := GPlayerMgr.GetPlayerByClientID(conn)
	if p == nil {
		return fmt.Errorf("(%v) (%d) ProcMsg GetPlayerByClientID nil", conn, msgID)
	}
	return f(p, msg[2:])
}

type PlayerMgr struct {
	playerIDMgr       map[string]*GamePlayer
	playerClientIDMgr map[kissnet.IConnection]*GamePlayer
	num               int64
	mutex             sync.RWMutex
}

func (mgr *PlayerMgr) GetPlayerByUserID(userID string) *GamePlayer {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()
	if v, ok := mgr.playerIDMgr[userID]; ok {
		return v
	}
	return nil
}
func (mgr *PlayerMgr) GetPlayerByClientID(conn kissnet.IConnection) *GamePlayer {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()
	if v, ok := mgr.playerClientIDMgr[conn]; ok {
		return v
	}
	return nil
}
func (mgr *PlayerMgr) AddPlayer(p *GamePlayer, conn kissnet.IConnection) error {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()
	mgr.playerClientIDMgr[p.Conn] = p
	mgr.playerIDMgr[p.UserID] = p
	mgr.num++
	return nil
}

func (mgr *PlayerMgr) DelPlayer(conn kissnet.IConnection) {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()
	p, ok := mgr.playerClientIDMgr[conn]
	if !ok {
		return
	}
	mgr.num--
	p.Dispose()

	delete(mgr.playerClientIDMgr, conn)
	delete(mgr.playerIDMgr, p.UserID)
}

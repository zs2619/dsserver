package gameplayer

import (
	"dsservices/gameserver/team"
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

// MarkDisconnected keeps the player in the registry (for reconnect) but flags it offline.
// If the player is in a realm, the entry is retained until the grace window expires.
func (mgr *PlayerMgr) MarkDisconnected(conn kissnet.IConnection, nowUnix int64) bool {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()
	p, ok := mgr.playerClientIDMgr[conn]
	if !ok {
		return false
	}
	// 重连后旧连接的异步回调可能晚到：只有当断开的连接仍是当前连接时才置空/标记掉线，
	// 否则会把刚重连上的新连接置空并把玩家误判为掉线被宽限回收。
	delete(mgr.playerClientIDMgr, conn)
	if p.Conn == conn {
		p.disconnectedAtUnix = nowUnix
		p.Conn = nil
	}
	return p.RealmID != ""
}

// SweepDisconnected removes players past the grace window and leaves their realms.
func (mgr *PlayerMgr) SweepDisconnected(nowUnix int64, graceSec int64) {
	mgr.mutex.Lock()
	var toRemove []*GamePlayer
	for uid, p := range mgr.playerIDMgr {
		if p.disconnectedAtUnix == 0 {
			continue
		}
		if nowUnix-p.disconnectedAtUnix >= graceSec {
			toRemove = append(toRemove, p)
			delete(mgr.playerIDMgr, uid)
		}
	}
	mgr.num -= int64(len(toRemove))
	mgr.mutex.Unlock()

	for _, p := range toRemove {
		if p.RealmID != "" {
			_, _ = team.QuitRealm(p.UserID, p.RealmID)
		} else if p.TeamID != "" {
			_, _ = team.QuitRealmTeam(p.UserID, p.TeamID)
		}
		p.Dispose()
	}
}

// CancelReconnect clears the disconnected flag on login.
func (mgr *PlayerMgr) CancelReconnect(userID string) {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()
	if p, ok := mgr.playerIDMgr[userID]; ok {
		p.disconnectedAtUnix = 0
	}
}

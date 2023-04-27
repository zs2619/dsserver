package dsa

import (
	"dsservices/dsaserver/proc"
	"dsservices/kissnet"
	"dsservices/pb"
	"bytes"
	"encoding/binary"
	"fmt"
	"sync"
	"time"
)

type DSInfoMgr struct {
	DSIDMgr   map[string]*DSInfo
	DSConnMgr map[kissnet.IConnection]*DSInfo
	num       int64
	mutex     sync.RWMutex
}

var GDSInfoMgr *DSInfoMgr = &DSInfoMgr{
	DSIDMgr:   make(map[string]*DSInfo),
	DSConnMgr: make(map[kissnet.IConnection]*DSInfo),
	num:       int64(0),
}

const (
	DS_Loading = iota
	DS_LoadOK
)
const (
	Realm_CreatIng = iota
	Realm_CreatOK
)

type RealmInfo struct {
	RealmState int
}

type DSInfo struct {
	Conn           kissnet.IConnection
	DsProcInfo     *proc.ProcInfo
	LastUpdateTime time.Time
	CurrTime       time.Time
	DSID           string ///ds 唯一id
	DSState        int
	RealmInfoMap   map[string]*RealmInfo
	RealmCfgID     string

	TeamIDPending string ///创建ds后，用此id创建副本
}

func (ds *DSInfo) SendMsg(msgID pb.DSA2DS_MsgID_MsgID, msg []byte) error {
	sendMsg := new(bytes.Buffer)
	binary.Write(sendMsg, binary.LittleEndian, uint16(msgID))
	sendMsg.Write(msg)
	if ds.Conn == nil {
		return fmt.Errorf("this.Conn == nil")
	}
	ds.Conn.SendMsg(sendMsg)
	return nil
}

func (ds *DSInfo) SetConnection(conn kissnet.IConnection) error {
	ds.Conn = conn
	GDSInfoMgr.DSConnMgr[ds.Conn] = ds
	return nil
}

func (dsMgr *DSInfoMgr) GetDSByID(dsID string) *DSInfo {
	dsMgr.mutex.RLock()
	defer dsMgr.mutex.RUnlock()
	if v, ok := dsMgr.DSIDMgr[dsID]; ok {
		return v
	}
	return nil
}
func (dsMgr *DSInfoMgr) GetDSByConn(conn kissnet.IConnection) *DSInfo {
	dsMgr.mutex.RLock()
	defer dsMgr.mutex.RUnlock()
	if v, ok := dsMgr.DSConnMgr[conn]; ok {
		return v
	}
	return nil
}

func (dsMgr *DSInfoMgr) AddDS(dsInfo *DSInfo) error {
	dsMgr.mutex.Lock()
	defer dsMgr.mutex.Unlock()
	dsMgr.DSIDMgr[dsInfo.DSID] = dsInfo
	dsMgr.num++
	return nil
}

func (dsMgr *DSInfoMgr) DelDSByConn(conn kissnet.IConnection) {
	dsMgr.mutex.Lock()
	defer dsMgr.mutex.Unlock()
	u, ok := dsMgr.DSConnMgr[conn]
	if !ok {
		return
	}
	dsMgr.num--

	delete(dsMgr.DSConnMgr, conn)
	delete(dsMgr.DSIDMgr, u.DSID)
}

func NewDS(dsID, realmCfgID, teamIDPending string) (dsInfo *DSInfo, err error) {
	procInfo, err := proc.StartProc(dsID, realmCfgID)
	if err != nil {
		return
	}
	dsInfo = &DSInfo{
		DsProcInfo:     procInfo,
		LastUpdateTime: time.Now(),
		CurrTime:       time.Now(),
		DSID:           dsID,
		DSState:        DS_Loading,
		RealmCfgID:     realmCfgID,
		Conn:           nil,
		TeamIDPending:  teamIDPending,
		RealmInfoMap:   make(map[string]*RealmInfo),
	}
	GDSInfoMgr.AddDS(dsInfo)
	return
}

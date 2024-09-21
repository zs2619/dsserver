package dsa

import (
	"bytes"
	"dsservices/dsaserver/proc"
	"dsservices/kissnet"
	"dsservices/pb"
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
	DS_CreatIng = iota
	DS_CreatOK
)

type DSInfo struct {
	DSConn         kissnet.IConnection
	DsProcInfo     *proc.ProcInfo
	LastUpdateTime time.Time
	CurrTime       time.Time
	DSID           string ///ds 唯一id
	DSState        int
	RealmCfgID     string
}

func NewDS(dsID, realmCfgID string) (dsInfo *DSInfo, err error) {
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
		DSConn:         nil,
	}
	GDSInfoMgr.AddDS(dsInfo)
	return
}
func (ds *DSInfo) KillDS() (err error) {
	if ds.DSConn != nil {
		ds.DSConn.Close()
	}
	if ds.DsProcInfo != nil {
		err = ds.DsProcInfo.KillProc()
	}
	return
}

func (ds *DSInfo) SendMsg(msgID pb.DSA2DS_MsgID_MsgID, msg []byte) error {
	sendMsg := new(bytes.Buffer)
	binary.Write(sendMsg, binary.LittleEndian, uint16(msgID))
	sendMsg.Write(msg)
	if ds.DSConn == nil {
		return fmt.Errorf("this.Conn == nil")
	}
	ds.DSConn.SendMsg(sendMsg)
	return nil
}

func (ds *DSInfo) SetConnection(conn kissnet.IConnection) error {
	ds.DSConn = conn
	GDSInfoMgr.DSConnMgr[ds.DSConn] = ds
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

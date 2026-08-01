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

	"github.com/sirupsen/logrus"
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
	DS_CreatIng = iota
	DS_CreatOK
)

type DSInfo struct {
	DSConn         kissnet.IConnection
	DsProcInfo     *proc.ProcInfo
	LastUpdateTime time.Time
	CurrTime       time.Time
	DSID           string
	RealmID        string
	DSState        int
	RealmCfgID     string
}

func NewDS(dsID, realmID, realmCfgID string) (dsInfo *DSInfo, err error) {
	procInfo, err := proc.StartProc(dsID, realmCfgID)
	if err != nil {
		return
	}
	dsInfo = &DSInfo{
		DsProcInfo:     procInfo,
		LastUpdateTime: time.Now(),
		CurrTime:       time.Now(),
		DSID:           dsID,
		RealmID:        realmID,
		DSState:        DS_CreatIng,
		RealmCfgID:     realmCfgID,
		DSConn:         nil,
	}
	GDSInfoMgr.AddDS(dsInfo)
	go watchDSProcess(dsInfo)
	return
}

// watchDSProcess is the sole Waiter for the DS OS process.
func watchDSProcess(ds *DSInfo) {
	if ds.DsProcInfo == nil || ds.DsProcInfo.OSProcess == nil {
		return
	}
	state, err := ds.DsProcInfo.OSProcess.Wait()
	if err != nil {
		logrus.WithError(err).WithField("dsID", ds.DSID).Warn("ds process wait error")
	} else {
		logrus.WithFields(logrus.Fields{"dsID": ds.DSID, "state": state}).Info("ds process exited")
	}
	ds = GDSInfoMgr.GetDSByID(ds.DSID)
	if ds == nil {
		return // already torn down (GAME_END / TCP disconnect / RemoveDS)
	}
	// Unexpected exit while still tracked: report DEAD so DSC releases capacity.
	if GReportClient != nil {
		reason := fmt.Sprintf("process exited before ready: %v", state)
		if ds.DSState != DS_CreatIng {
			reason = fmt.Sprintf("process exited: %v", state)
		}
		GReportClient.ReportDSEvent(&pb.ReportDSEventReq{
			DsID:       ds.DSID,
			RealmID:    ds.RealmID,
			EventType:  pb.DSEventType_DS_EVENT_DEAD,
			FailReason: reason,
		})
	}
	GDSInfoMgr.DelDSByID(ds.DSID)
}

// KillDS signals process/conn stop. Port release + Wait live in DelDSByID / watchDSProcess.
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
	if conn == nil {
		return fmt.Errorf("conn nil")
	}
	ds.DSConn = conn
	GDSInfoMgr.DSConnMgr[ds.DSConn] = ds
	return nil
}

func (dsMgr *DSInfoMgr) Count() int {
	dsMgr.mutex.RLock()
	defer dsMgr.mutex.RUnlock()
	return int(dsMgr.num)
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

func (dsMgr *DSInfoMgr) DelDSByID(dsID string) {
	dsMgr.mutex.Lock()
	defer dsMgr.mutex.Unlock()
	ds, ok := dsMgr.DSIDMgr[dsID]
	if !ok {
		return
	}
	dsMgr.num--
	delete(dsMgr.DSIDMgr, dsID)
	if ds.DSConn != nil {
		delete(dsMgr.DSConnMgr, ds.DSConn)
	}
	// Sole port release point; zero Port so a later path cannot double-release.
	if ds.DsProcInfo != nil && ds.DsProcInfo.Port > 0 {
		proc.GPortMgr.ReleasePort(ds.DsProcInfo.Port)
		ds.DsProcInfo.Port = 0
	}
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

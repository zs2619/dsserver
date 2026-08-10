package dsa

import (
	"bytes"
	"dsservices/dsaserver/proc"
	"github.com/zs2619/kissnet-go"
	"dsservices/pb"
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
)

type DSInfoMgr struct {
	DSIDMgr   map[string]*DSInfo
	DSConnMgr map[kissnet.Connection]*DSInfo
	num       int64
	mutex     sync.RWMutex
}

var GDSInfoMgr *DSInfoMgr = &DSInfoMgr{
	DSIDMgr:   make(map[string]*DSInfo),
	DSConnMgr: make(map[kissnet.Connection]*DSInfo),
	num:       int64(0),
}

const (
	DS_CreatIng = iota
	DS_CreatOK
)

// GAdvertiseIP 是玩家可达 IP（container=status.hostIP，process=config dsIP），由 main.go 注入。
var GAdvertiseIP string

type DSInfo struct {
	DSConn         kissnet.Connection
	Port           int // PortMgr 分配的游戏口（Fork 后无 OSProcess，口直接存此）
	LastUpdateTime time.Time
	CurrTime       time.Time
	DSID           string
	RealmID        string
	DSState        int
	RealmCfgID     string
	PendingStop    bool // RemoveDS 时实例尚未 CreateOK：置 true，CreateOK 到即发 StopDS+Del
}

// NewDS 分配端口并请求模板 fork 一个实例。不再 os.StartProcess。
func NewDS(dsID, realmID, realmCfgID string) (dsInfo *DSInfo, err error) {
	port := proc.GPortMgr.GetValidPort()
	if port <= 0 {
		return nil, fmt.Errorf("no available port")
	}
	dsInfo = &DSInfo{
		Port:           port,
		LastUpdateTime: time.Now(),
		CurrTime:       time.Now(),
		DSID:           dsID,
		RealmID:        realmID,
		DSState:        DS_CreatIng,
		RealmCfgID:     realmCfgID,
		DSConn:         nil,
	}
	GDSInfoMgr.AddDS(dsInfo)
	// GTemplateClient 全局非 nil；须查 IsReady（模板未连/已断）。
	if GTemplateClient == nil || !GTemplateClient.IsReady() {
		GDSInfoMgr.DelDSByID(dsID)
		return nil, fmt.Errorf("template client not ready")
	}
	if err = GTemplateClient.ForkInstance(dsID, realmID, realmCfgID, port, GAdvertiseIP); err != nil {
		// DelDSByID 统一释放口，勿在此重复 ReleasePort。
		GDSInfoMgr.DelDSByID(dsID)
		return nil, err
	}
	return
}

// KillDS 发 StopDS 让实例优雅退出；并关连接。端口释放集中在 DelDSByID。
func (ds *DSInfo) KillDS() (err error) {
	if ds.DSConn != nil {
		_ = ds.SendMsg(pb.DSA2DS_MsgID_StopDS, mustMarshalStopDS(ds.DSID))
		ds.DSConn.Close()
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
	return ds.DSConn.SendMsg(sendMsg)
}

func (ds *DSInfo) SetConnection(conn kissnet.Connection) error {
	return GDSInfoMgr.BindConn(ds, conn)
}

// BindConn 在锁下登记 DS 的 TCP 连接。
func (m *DSInfoMgr) BindConn(ds *DSInfo, conn kissnet.Connection) error {
	if ds == nil || conn == nil {
		return fmt.Errorf("ds/conn nil")
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if _, ok := m.DSIDMgr[ds.DSID]; !ok {
		return fmt.Errorf("ds not tracked")
	}
	if ds.DSConn != nil && ds.DSConn != conn {
		delete(m.DSConnMgr, ds.DSConn)
	}
	ds.DSConn = conn
	m.DSConnMgr[conn] = ds
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

func (dsMgr *DSInfoMgr) GetDSByConn(conn kissnet.Connection) *DSInfo {
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
	if ds.Port > 0 {
		proc.GPortMgr.ReleasePort(ds.Port)
		ds.Port = 0
	}
}

func (dsMgr *DSInfoMgr) DelDSByConn(conn kissnet.Connection) {
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

// ShutdownAll 退出时杀掉全部 DS：先报 DEAD 再摘表再 Kill。
// 可重复调用；第二次为空操作。
func (dsMgr *DSInfoMgr) ShutdownAll() {
	dsMgr.mutex.Lock()
	list := make([]*DSInfo, 0, len(dsMgr.DSIDMgr))
	for _, ds := range dsMgr.DSIDMgr {
		list = append(list, ds)
	}
	dsMgr.mutex.Unlock()

	for _, ds := range list {
		logrus.WithField("dsID", ds.DSID).Info("shutdown: killing DS")
		if GReportClient != nil {
			GReportClient.ReportDSEvent(&pb.ReportDSEventReq{
				DsID:       ds.DSID,
				RealmID:    ds.RealmID,
				EventType:  pb.DSEventType_DS_EVENT_DEAD,
				FailReason: "dsa shutdown",
			})
		}
		// 先摘表，避免 Kill 触发的 OnDisconnect 再报一次 DEAD。
		dsMgr.DelDSByID(ds.DSID)
		if err := ds.KillDS(); err != nil {
			logrus.WithError(err).WithField("dsID", ds.DSID).Warn("shutdown KillDS")
		}
	}
}

// SweepStaleCreating 回收 fork 后始终未 CreateOK 的 DS（实例进程死/连不上且未报 ForkFailed）。
// 必须在 EventLoop 上调用（与 dsCreateOK/OnDisconnect 串行，无竞态）。
// 超时后 DelDSByID（释放 PortMgr 口）并上报 DEAD，DSC 侧 realm 走 fail。
func (dsMgr *DSInfoMgr) SweepStaleCreating(timeoutSec int64) {
	if timeoutSec <= 0 {
		return
	}
	now := time.Now()
	dsMgr.mutex.Lock()
	list := make([]*DSInfo, 0)
	for _, ds := range dsMgr.DSIDMgr {
		if ds.DSState == DS_CreatIng && now.Sub(ds.CurrTime) > time.Duration(timeoutSec)*time.Second {
			list = append(list, ds)
		}
	}
	dsMgr.mutex.Unlock()

	for _, ds := range list {
		logrus.WithFields(logrus.Fields{"dsID": ds.DSID, "ageSec": now.Sub(ds.CurrTime).Seconds()}).Warn("ds creating timeout (no CreateOK), reclaiming")
		if GReportClient != nil {
			GReportClient.ReportDSEvent(&pb.ReportDSEventReq{
				DsID:       ds.DSID,
				RealmID:    ds.RealmID,
				EventType:  pb.DSEventType_DS_EVENT_DEAD,
				FailReason: "dsa creating timeout (no CreateOK)",
			})
		}
		dsMgr.DelDSByID(ds.DSID)
	}
}

func mustMarshalStopDS(dsID string) []byte {
	b, _ := proto.Marshal(&pb.DSA2DS_StopDSReq{DsID: dsID})
	return b
}

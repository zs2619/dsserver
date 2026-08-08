package dsa

import (
	"bytes"
	"dsservices/kissnet"
	"dsservices/pb"
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

// TemplateClient 维护与 DS 模板进程的连接，承担 fork 实例的请求。
// 模板进程连上 DSA 的模板口后发 TemplateReady，DSA 据此标记 ready 并向 DSC 上报。
type TemplateClient struct {
	mutex  sync.Mutex
	conn   kissnet.Connection
	ready  bool
}

var GTemplateClient *TemplateClient = &TemplateClient{}

// TemplateCB 是模板口的连接回调，独立于实例口的 DSAServerCB。
type TemplateCB struct{}

var GTemplateCB TemplateCB

func (c *TemplateCB) OnConnectCB(conn kissnet.Connection) error {
	logrus.Info("template OnConnect")
	return nil
}

func (c *TemplateCB) OnDisconnectCB(conn kissnet.Connection) error {
	logrus.Warn("template OnDisconnect")
	GTemplateClient.setConn(nil, false)
	// 同步 ReportClient：模板断开→心跳标 false→Lua 选机跳过本 Agent，避免调度却 fork 不了。
	if GReportClient != nil {
		GReportClient.SetTemplateReady(false)
	}
	return nil
}

func (c *TemplateCB) OnMsgCB(conn kissnet.Connection, msg []byte) error {
	if len(msg) < 2 {
		return fmt.Errorf("template msg len error")
	}
	msgID := pb.DS2DSA_MsgID_MsgID(binary.LittleEndian.Uint16(msg))
	body := msg[2:]
	switch msgID {
	case pb.DS2DSA_MsgID_TemplateReady:
		GTemplateClient.setConn(conn, true)
		logrus.Info("template TemplateReady")
		if GReportClient != nil {
			GReportClient.SetTemplateReady(true)
		}
	case pb.DS2DSA_MsgID_ForkFailed:
		handleForkFailed(body)
	default:
		return fmt.Errorf("unexpected template msgID %d", msgID)
	}
	return nil
}

func (tc *TemplateClient) setConn(conn kissnet.Connection, ready bool) {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()
	tc.conn = conn
	tc.ready = ready
}

// IsReady 返回模板是否就绪。
func (tc *TemplateClient) IsReady() bool {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()
	return tc.ready && tc.conn != nil
}

// ForkInstance 在模板口连接上发送 fork 请求。
func (tc *TemplateClient) ForkInstance(dsID, realmID, realmCfgID string, port int, advertiseIP string) error {
	tc.mutex.Lock()
	conn := tc.conn
	ready := tc.ready
	tc.mutex.Unlock()
	if !ready || conn == nil {
		return fmt.Errorf("template not ready")
	}
	req := &pb.DSA2DS_ForkInstanceReq{
		DsID:       dsID,
		RealmID:    realmID,
		RealmCfgID: realmCfgID,
		Port:       int32(port),
		AdvertiseIP: advertiseIP,
	}
	body, err := proto.Marshal(req)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint16(pb.DSA2DS_MsgID_ForkInstance))
	buf.Write(body)
	return conn.SendMsg(buf)
}

// handleForkFailed: 模板 fork 失败，按 DEAD 上报并摘表（释放口）。
func handleForkFailed(body []byte) {
	req := &pb.DS2DSA_ForkFailed{}
	if err := proto.Unmarshal(body, req); err != nil {
		logrus.WithError(err).Error("ForkFailed unmarshal")
		return
	}
	logrus.WithFields(logrus.Fields{"dsID": req.DsID, "reason": req.Reason}).Warn("template ForkFailed")
	ds := GDSInfoMgr.GetDSByID(req.DsID)
	if ds == nil {
		return
	}
	if GReportClient != nil {
		GReportClient.ReportDSEvent(&pb.ReportDSEventReq{
			DsID:       ds.DSID,
			RealmID:    ds.RealmID,
			EventType:  pb.DSEventType_DS_EVENT_DEAD,
			FailReason: "fork failed: " + req.Reason,
		})
	}
	GDSInfoMgr.DelDSByID(ds.DSID)
}

package dsa

import (
	"dsservices/kissnet"
	"dsservices/pb"
	"encoding/binary"
	"fmt"

	"github.com/sirupsen/logrus"
)

type DSAServerCB struct {
}

var GDSAServerCB DSAServerCB

func (dsaCB *DSAServerCB) OnDisconnectCB(conn kissnet.Connection) error {
	logrus.WithFields(logrus.Fields{}).Info("OnDisconnectCB")
	ds := GDSInfoMgr.GetDSByConn(conn)
	if ds == nil {
		GDSInfoMgr.DelDSByConn(conn)
		return nil
	}
	// TCP 断开：上报 DEAD，释放 Redis 容量与 realm；避免仍 ready。
	if GReportClient != nil {
		reason := "ds tcp disconnected"
		if ds.DSState == DS_CreatIng {
			reason = "ds disconnected before CreateOK"
		}
		GReportClient.ReportDSEvent(&pb.ReportDSEventReq{
			DsID:       ds.DSID,
			RealmID:    ds.RealmID,
			EventType:  pb.DSEventType_DS_EVENT_DEAD,
			FailReason: reason,
		})
	}
	// 先摘表（DelDSByID 依赖 ds.DSConn 清 DSConnMgr，故此处不能先置 nil），
	// 再置 nil DSConn 防 KillDS.Close 二次进本回调重复报 DEAD。
	GDSInfoMgr.DelDSByID(ds.DSID)
	ds.DSConn = nil
	_ = ds.KillDS()
	return nil
}

func (dsaCB *DSAServerCB) OnConnectCB(conn kissnet.Connection) error {
	logrus.WithFields(logrus.Fields{}).Info("OnConnectCB")
	return nil
}

func (dsaCB *DSAServerCB) OnMsgCB(conn kissnet.Connection, msg []byte) error {
	if msg == nil {
		return nil
	}
	if len(msg) < 2 {
		return fmt.Errorf("dsa msg len error")
	}

	msgID := pb.DS2DSA_MsgID_MsgID(binary.LittleEndian.Uint16(msg))
	err := ProcDSMsg(conn, msgID, msg[2:])
	if err != nil {
		return nil
	}

	return nil
}
func ProcDSMsg(conn kissnet.Connection, msgID pb.DS2DSA_MsgID_MsgID, msg []byte) error {
	if msgID == pb.DS2DSA_MsgID_DSCreateOK {
		err := dsCreateOK(conn, msg)
		return err
	} else {
		f, ok := msgMap[msgID]
		if !ok {
			return fmt.Errorf("(%d) msgID nil", msgID)
		}
		u := GDSInfoMgr.GetDSByConn(conn)
		if u == nil {
			return fmt.Errorf("(%d) ProcMsg ds nil", msgID)
		}
		return f(u, msg)
	}
}

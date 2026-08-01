package dsa

import (
	"dsservices/kissnet"
	"dsservices/pb"
	"fmt"

	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

type FuncMsg func(*DSInfo, []byte) error

var (
	msgMap map[pb.DS2DSA_MsgID_MsgID]FuncMsg = make(map[pb.DS2DSA_MsgID_MsgID]FuncMsg)
)

func init() {
	RegisterMsgMapFunc()
}

func RegisterMsgMapFunc() {
	msgMap[pb.DS2DSA_MsgID_DSGameEnd] = dsGameEnd
	msgMap[pb.DS2DSA_MsgID_DSUpdateState] = dsUpdateState
}

func dsCreateOK(conn kissnet.IConnection, msg []byte) error {
	req := &pb.DS2DSA_CreateOK{}
	err := proto.Unmarshal(msg, req)
	if err != nil {
		return err
	}
	dsInfo := GDSInfoMgr.GetDSByID(req.DsID)
	if dsInfo == nil {
		return nil
	}
	dsInfo.SetConnection(conn)
	if dsInfo.DSState == DS_CreatIng {
		dsInfo.DSState = DS_CreatOK
		addr := fmt.Sprintf("%s:%d", dsInfo.DsProcInfo.Ip, dsInfo.DsProcInfo.Port)
		logrus.WithFields(logrus.Fields{"dsID": dsInfo.DSID, "dsAddr": addr}).Info("DS CreateOK")
		if GReportClient != nil {
			GReportClient.ReportDSEvent(&pb.ReportDSEventReq{
				DsID:      dsInfo.DSID,
				RealmID:   dsInfo.RealmID,
				EventType: pb.DSEventType_DS_EVENT_READY,
				DsAddr:    addr,
			})
		}
	}
	return nil
}

func dsGameEnd(dsInfo *DSInfo, msg []byte) error {
	if GReportClient != nil {
		GReportClient.ReportDSEvent(&pb.ReportDSEventReq{
			DsID:      dsInfo.DSID,
			RealmID:   dsInfo.RealmID,
			EventType: pb.DSEventType_DS_EVENT_GAME_END,
		})
	}
	// Untrack first so watchDSProcess Wait does not report DEAD after intentional end.
	GDSInfoMgr.DelDSByID(dsInfo.DSID)
	_ = dsInfo.KillDS()
	return nil
}

func dsUpdateState(dsInfo *DSInfo, msg []byte) error {
	if GReportClient != nil {
		GReportClient.ReportDSEvent(&pb.ReportDSEventReq{
			DsID:      dsInfo.DSID,
			RealmID:   dsInfo.RealmID,
			EventType: pb.DSEventType_DS_EVENT_PLAYER_ENTER,
		})
	}
	return nil
}

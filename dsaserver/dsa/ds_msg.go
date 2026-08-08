package dsa

import (
	"dsservices/dsaserver/proc"
	"dsservices/kissnet"
	"dsservices/pb"
	"fmt"
	"strconv"
	"strings"

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

func dsCreateOK(conn kissnet.Connection, msg []byte) error {
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
	// RemoveDS 在实例尚未 CreateOK 时只置 PendingStop；CreateOK 到即不升级 ready，直接 Stop+Del（防孤儿）。
	if dsInfo.PendingStop {
		dsInfo.PendingStop = false
		logrus.WithField("dsID", dsInfo.DSID).Info("DS CreateOK hit pendingStop, stopping")
		GDSInfoMgr.DelDSByID(dsInfo.DSID)
		_ = dsInfo.KillDS()
		return nil
	}
	if dsInfo.DSState == DS_CreatIng {
		dsInfo.DSState = DS_CreatOK
		// DS 内可能换口：CreateOK 报最终口。校正记账——释放推荐口，标记实际口。
		addr := req.DsAddr
		if actualPort := parsePort(addr); actualPort > 0 && actualPort != dsInfo.Port {
			proc.GPortMgr.ReleasePort(dsInfo.Port)
			if proc.GPortMgr.MarkUsed(actualPort) {
				dsInfo.Port = actualPort
			}
		}
		if addr == "" {
			ip := GAdvertiseIP
			if ip == "" {
				ip = "127.0.0.1"
			}
			addr = fmt.Sprintf("%s:%d", ip, dsInfo.Port)
		}
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
	// Untrack first so OnDisconnect does not report DEAD after intentional end.
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

// parsePort 从 "ip:port" 中解析端口；失败返回 0。
func parsePort(addr string) int {
	idx := strings.LastIndex(addr, ":")
	if idx < 0 {
		return 0
	}
	n, err := strconv.Atoi(addr[idx+1:])
	if err != nil {
		return 0
	}
	return n
}

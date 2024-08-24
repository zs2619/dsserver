package dsa

import (
	"dsservices/kissnet"
	"dsservices/pb"
	"fmt"
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
	msgMap[pb.DS2DSA_MsgID_DSGameEnd] = dsGameEndResp
	msgMap[pb.DS2DSA_MsgID_DSRealmCreateOk] = dsRealmCreateOkResp
}

func dsLoadOK(conn kissnet.IConnection, msg []byte) error {
	req := &pb.DSLoadOKReq{}
	err := proto.Unmarshal(msg, req)
	if err != nil {
		return err
	}
	dsInfo := GDSInfoMgr.GetDSByID(req.DsID)
	if dsInfo == nil {
		return nil
	}
	dsInfo.SetConnection(conn)
	if dsInfo.DSState == DS_Loading {
		/// 创建房间
		dsInfo.DSState = DS_LoadOK

		dsInfo.RealmInfoMap[dsInfo.TeamIDPending] = &RealmInfo{RealmState: Realm_CreatIng}
		dsRealmCreate(dsInfo, dsInfo.TeamIDPending)
		dsInfo.TeamIDPending = ""

	}
	return nil
}

func dsGameEndResp(dsInfo *DSInfo, msg []byte) error {
	//TODO回收ds进程

	return nil
}

func dsRealmCreateOkResp(dsInfo *DSInfo, msg []byte) error {
	logrus.Debug("dsRealmCreateOk")
	resp := &pb.RealmCreateResp{}
	err := proto.Unmarshal(msg, resp)
	if err != nil {
		return err
	}

	realID := "{team}_" + strings.ToLower(resp.RealmID)
	addr := fmt.Sprintf("%s:%d", dsInfo.DsProcInfo.Ip, dsInfo.DsProcInfo.Port)
	rpcResp := &pb.RpcCreateRealmResult{
		DsAddr:     addr,
		TeamID:     realID,
		RealmCfgID: dsInfo.RealmCfgID,
	}
	_, ok := dsInfo.RealmInfoMap[resp.RealmID]
	if ok {
		dsInfo.RealmInfoMap[resp.RealmID].RealmState = Realm_CreatOK
	}
	GDSAClient.SendStreamService(rpcResp)
	return nil
}

func dsRealmCreate(dsInfo *DSInfo, realmID string) error {
	RealmCreateReq := &pb.RealmCreateReq{}
	//TODo:暂时去掉
	RealmCreateReq.RealmId = realmID[7:]
	MsgData, err := proto.Marshal(RealmCreateReq)
	if err != nil {
		return err
	}
	dsInfo.SendMsg(pb.DSA2DS_MsgID_CreateRealm, MsgData)
	return nil
}

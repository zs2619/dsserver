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
	msgMap[pb.DS2DSA_MsgID_DSGameEnd] = dsGameEndResp
	msgMap[pb.DS2DSA_MsgID_DSDSCreateOk] = dsDSCreateOkResp
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

		dsDSCreate(dsInfo)

	}
	return nil
}

func dsGameEndResp(dsInfo *DSInfo, msg []byte) error {
	//TODO回收ds进程

	return nil
}

func dsDSCreateOkResp(dsInfo *DSInfo, msg []byte) error {
	logrus.Debug("dsDSCreateOk")
	resp := &pb.DSCreateResp{}
	err := proto.Unmarshal(msg, resp)
	if err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", dsInfo.DsProcInfo.Ip, dsInfo.DsProcInfo.Port)
	rpcResp := &pb.RpcCreateDSResult{
		DsAddr:     addr,
		RealmCfgID: dsInfo.RealmCfgID,
	}
	GDSAClient.Send2DSC(rpcResp)
	return nil
}

func dsDSCreate(dsInfo *DSInfo) error {
	DSCreateReq := &pb.DSCreateReq{}
	DSCreateReq.DsID = dsInfo.DSID
	MsgData, err := proto.Marshal(DSCreateReq)
	if err != nil {
		return err
	}
	dsInfo.SendMsg(pb.DSA2DS_MsgID_CreateDS, MsgData)
	return nil
}

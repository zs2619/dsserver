package dsa

import (
	"dsservices/kissnet"
	"dsservices/pb"
	"fmt"

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
		/// 创建DS成功
		dsInfo.DSState = DS_CreatOK

		addr := fmt.Sprintf("%s:%d", dsInfo.DsProcInfo.Ip, dsInfo.DsProcInfo.Port)
		rpcResp := &pb.StreamCreateDSResult{
			DsAddr:     addr,
			RealmCfgID: dsInfo.RealmCfgID,
		}
		GDSAClient.Send2DSC(rpcResp)
	}
	return nil
}

func dsGameEnd(dsInfo *DSInfo, msg []byte) error {
	//TODO回收ds进程
	return nil
}

func dsUpdateState(dsInfo *DSInfo, msg []byte) error {
	return nil
}

package gamedsc

import (
	"dsservices/dscserver/dasdsc"
	"dsservices/dscserver/realm"
	"dsservices/pb"
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RPCGameDscServer struct {
	pb.UnimplementedGameDscRealmServer
}

func (gdServer *RPCGameDscServer) CreateRealm(context.Context, *pb.RpcCreateRealmReq) (*pb.RpcCreateRealmResp, error) {

	return nil, nil
}
func (gdServer *RPCGameDscServer) JoinRealm(context.Context, *pb.RpcJoinRealmReq) (*pb.RpcJoinRealmResp, error) {

	return nil, nil
}
func (gdServer *RPCGameDscServer) QuickJoinRealm(ctx context.Context, req *pb.RpcQuickJoinRealmReq) (*pb.RpcQuickJoinRealmResp, error) {
	var dsAddr string
	pending := false
	realmInfo, err := realm.GetRealmInfo(req.RealmCfgID)
	if err != nil && realmInfo == nil {
		//创建
		pending = true
		dasdsc.CreateRealm(primitive.NewObjectID().Hex(), req.TeamID, req.RealmCfgID)
	} else {
		dsAddr = realmInfo.Addr
	}

	return &pb.RpcQuickJoinRealmResp{
		Header:     &pb.RpcHeader{Ret: pb.RpcHeader_OK},
		DsAddr:     dsAddr,
		Pending:    pending,
		RealmCfgID: req.RealmCfgID,
		TeamID:     req.TeamID}, nil
}
func (gdServer *RPCGameDscServer) QueryRealmList(context.Context, *pb.RpcQueryRealmListRealmReq) (*pb.RpcQueryRealmListRealmResp, error) {

	return nil, nil
}

func (gdServer *RPCGameDscServer) NotifyJoinRealm(realm pb.GameDscRealm_NotifyJoinRealmServer) error {
	for info := range dasdsc.CreateRealmResult {
		if info == nil {
			break
		}
		notifyInfo := &pb.NotifyJoinRealmResp{
			Header:     &pb.RpcHeader{Ret: pb.RpcHeader_OK},
			TeamID:     info.TeamID,
			RealmCfgID: info.RealmCfgID,
			DsAddr:     info.DsAddr,
		}
		realm.Send(notifyInfo)
	}
	return nil
}

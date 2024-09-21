package gamedsc

import (
	"context"
	"dsservices/dscserver/dsamgr"
	"dsservices/pb"
)

type RPCGameDscServer struct {
	pb.UnimplementedGameDscDSServer
}

func (gdServer *RPCGameDscServer) CreateRealm(ctx context.Context, createRealReq *pb.RpcCreateDSReq) (*pb.RpcCreateDSResp, error) {
	dsaInfo := dsamgr.GDSAMgr.GetOptiamlDsa()
	if dsaInfo == nil {
		//Todo 没有资源
		return nil, nil
	}
	dsaInfo.AddDS(createRealReq.RealmCfgID)
	return nil, nil
}

func (gdServer *RPCGameDscServer) RemoveRealm(ctx context.Context, req *pb.RpcRemoveDSReq) (*pb.RpcRemoveDSResp, error) {
	dsaInfo := dsamgr.GDSAMgr.GetDsaByID(req.DsID)
	if dsaInfo == nil {
		return nil, nil
	}
	dsaInfo.AddDS(req.RealmCfgID)
	return nil, nil
}

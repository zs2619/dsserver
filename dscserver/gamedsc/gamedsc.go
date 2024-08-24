package gamedsc

import (
	"context"
	"dsservices/pb"
)

type RPCGameDscServer struct {
	pb.UnimplementedGameDscRealmServer
}

func (gdServer *RPCGameDscServer) CreateRealm(context.Context, *pb.RpcCreateRealmReq) (*pb.RpcCreateRealmResp, error) {

	return nil, nil
}

func (gdServer *RPCGameDscServer) RemoveRealm(context.Context, *pb.RpcCreateRealmReq) (*pb.RpcCreateRealmResp, error) {
	return nil, nil
}

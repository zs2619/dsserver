package dsa

import (
	"context"
	"dsservices/pb"
	"fmt"
	"net"
	"sync"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type AgentServer struct {
	pb.UnimplementedRpcDsaAgentServer
	port       int
	grpcServer *grpc.Server
	stopOnce   sync.Once
}

func NewAgentServer(port int) *AgentServer {
	return &AgentServer{
		port:       port,
		grpcServer: grpc.NewServer(),
	}
}

func (s *AgentServer) Run() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return err
	}
	pb.RegisterRpcDsaAgentServer(s.grpcServer, s)
	logrus.WithField("port", s.port).Info("dsa agent grpc serve")
	return s.grpcServer.Serve(lis)
}

func (s *AgentServer) Stop() {
	s.stopOnce.Do(func() {
		if s.grpcServer != nil {
			s.grpcServer.GracefulStop()
		}
	})
}

func (s *AgentServer) CreateDS(ctx context.Context, req *pb.DsaCreateDSReq) (*pb.DsaCreateDSResp, error) {
	if req.DsID == "" || req.RealmCfgID == "" {
		return &pb.DsaCreateDSResp{Header: &pb.RpcHeader{Ret: pb.RpcHeader_ReqParamError, Errmsg: "dsID/realmCfgID required"}}, nil
	}
	if existing := GDSInfoMgr.GetDSByID(req.DsID); existing != nil {
		// 已存在：Fork 后无真实地址，返回空 dsAddr（真地址等 CreateOK 上报）。
		return &pb.DsaCreateDSResp{Header: &pb.RpcHeader{Ret: pb.RpcHeader_OK}, DsAddr: ""}, nil
	}
	dsInfo, err := NewDS(req.DsID, req.RealmID, req.RealmCfgID)
	if err != nil {
		logrus.WithError(err).WithField("dsID", req.DsID).Error("NewDS")
		return &pb.DsaCreateDSResp{Header: &pb.RpcHeader{Ret: pb.RpcHeader_ReqParamError, Errmsg: err.Error()}}, nil
	}
	// Fork 后无 OSProcess，DsAddr 由实例 CreateOK 上报；此处返回空占位。
	logrus.WithFields(logrus.Fields{"dsID": req.DsID, "realmID": req.RealmID, "port": dsInfo.Port}).Info("CreateDS forked, waiting CreateOK")
	return &pb.DsaCreateDSResp{Header: &pb.RpcHeader{Ret: pb.RpcHeader_OK}, DsAddr: ""}, nil
}

func (s *AgentServer) RemoveDS(ctx context.Context, req *pb.DsaRemoveDSReq) (*pb.DsaRemoveDSResp, error) {
	if req.DsID == "" {
		return &pb.DsaRemoveDSResp{Header: &pb.RpcHeader{Ret: pb.RpcHeader_ReqParamError}}, nil
	}
	ds := GDSInfoMgr.GetDSByID(req.DsID)
	if ds != nil {
		if ds.DSState == DS_CreatIng {
			// 实例尚未 CreateOK：置 PendingStop，等 CreateOK 到再 Stop+Del（避免孤儿实例）。
			ds.PendingStop = true
		} else {
			GDSInfoMgr.DelDSByID(req.DsID)
			_ = ds.KillDS()
		}
	}
	return &pb.DsaRemoveDSResp{Header: &pb.RpcHeader{Ret: pb.RpcHeader_OK}}, nil
}

func (s *AgentServer) GetStatus(ctx context.Context, req *pb.DsaGetStatusReq) (*pb.DsaGetStatusResp, error) {
	return &pb.DsaGetStatusResp{
		Header:  &pb.RpcHeader{Ret: pb.RpcHeader_OK},
		DsCount: int32(GDSInfoMgr.Count()),
	}, nil
}

package dsa

import (
	"context"
	"dsservices/pb"
	"fmt"
	"net"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type AgentServer struct {
	pb.UnimplementedRpcDsaAgentServer
	port       int
	grpcServer *grpc.Server
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
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}
}

func (s *AgentServer) CreateDS(ctx context.Context, req *pb.DsaCreateDSReq) (*pb.DsaCreateDSResp, error) {
	if req.DsID == "" || req.RealmCfgID == "" {
		return &pb.DsaCreateDSResp{Header: &pb.RpcHeader{Ret: pb.RpcHeader_ReqParamError, Errmsg: "dsID/realmCfgID required"}}, nil
	}
	if existing := GDSInfoMgr.GetDSByID(req.DsID); existing != nil {
		addr := ""
		if existing.DsProcInfo != nil {
			addr = fmt.Sprintf("%s:%d", existing.DsProcInfo.Ip, existing.DsProcInfo.Port)
		}
		return &pb.DsaCreateDSResp{Header: &pb.RpcHeader{Ret: pb.RpcHeader_OK}, DsAddr: addr}, nil
	}
	dsInfo, err := NewDS(req.DsID, req.RealmID, req.RealmCfgID)
	if err != nil {
		logrus.WithError(err).WithField("dsID", req.DsID).Error("NewDS")
		return &pb.DsaCreateDSResp{Header: &pb.RpcHeader{Ret: pb.RpcHeader_ReqParamError, Errmsg: err.Error()}}, nil
	}
	// 仅分配端口并拉起进程；realm ready 等 DS TCP CreateOK 后再 ReportDSEvent。
	addr := fmt.Sprintf("%s:%d", dsInfo.DsProcInfo.Ip, dsInfo.DsProcInfo.Port)
	logrus.WithFields(logrus.Fields{"dsID": req.DsID, "realmID": req.RealmID, "dsAddr": addr}).Info("CreateDS started, waiting CreateOK")
	return &pb.DsaCreateDSResp{Header: &pb.RpcHeader{Ret: pb.RpcHeader_OK}, DsAddr: addr}, nil
}

func (s *AgentServer) RemoveDS(ctx context.Context, req *pb.DsaRemoveDSReq) (*pb.DsaRemoveDSResp, error) {
	if req.DsID == "" {
		return &pb.DsaRemoveDSResp{Header: &pb.RpcHeader{Ret: pb.RpcHeader_ReqParamError}}, nil
	}
	ds := GDSInfoMgr.GetDSByID(req.DsID)
	if ds != nil {
		GDSInfoMgr.DelDSByID(req.DsID)
		_ = ds.KillDS()
	}
	return &pb.DsaRemoveDSResp{Header: &pb.RpcHeader{Ret: pb.RpcHeader_OK}}, nil
}

func (s *AgentServer) GetStatus(ctx context.Context, req *pb.DsaGetStatusReq) (*pb.DsaGetStatusResp, error) {
	return &pb.DsaGetStatusResp{
		Header:  &pb.RpcHeader{Ret: pb.RpcHeader_OK},
		DsCount: int32(GDSInfoMgr.Count()),
	}, nil
}

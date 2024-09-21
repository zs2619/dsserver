package dsc

import (
	"dsservices/dscserver/dasdsc"
	"dsservices/dscserver/gamedsc"
	"dsservices/pb"
	"fmt"
	"net"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type DSCServer struct {
	port       int
	grpcServer *grpc.Server

	streamServerEventChan chan *pb.StreamServerEvent
}

var GDSCServer *DSCServer

func (dscServer *DSCServer) Run() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", dscServer.port))
	if err != nil {
		logrus.Fatalf("failed to listen: %v", err)
	}
	pb.RegisterRpcGameDscServer(dscServer.grpcServer, &gamedsc.RPCGameDscServer{})
	pb.RegisterStreamDscDSAServer(dscServer.grpcServer, &dasdsc.RPCDasDscServer{})
	dscServer.grpcServer.Serve(lis)
}

func NewDSCServer(port int) (dscServer *DSCServer, err error) {
	var opts []grpc.ServerOption
	dscServer = &DSCServer{
		port: port,
	}
	dscServer.grpcServer = grpc.NewServer(opts...)
	return
}

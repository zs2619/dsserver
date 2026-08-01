package dsc

import (
	"dsservices/dscserver/dsareport"
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
}

var GDSCServer *DSCServer

func (dscServer *DSCServer) Run() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", dscServer.port))
	if err != nil {
		logrus.WithError(err).Fatalf("failed to listen")
	}
	pb.RegisterRpcGameDscServer(dscServer.grpcServer, &gamedsc.RPCGameDscServer{})
	pb.RegisterRpcDsaReportServer(dscServer.grpcServer, &dsareport.Server{})
	logrus.WithField("port", dscServer.port).Info("dsc grpc serve")
	dscServer.grpcServer.Serve(lis)
}

func NewDSCServer(port int) (dscServer *DSCServer, err error) {
	dscServer = &DSCServer{
		port:       port,
		grpcServer: grpc.NewServer(),
	}
	GDSCServer = dscServer
	return
}

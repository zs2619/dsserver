package dsc

import (
	"dsservices/dscserver/dsareport"
	"dsservices/dscserver/gamedsc"
	"dsservices/pb"
	"fmt"
	"net"
	"sync"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type DSCServer struct {
	port       int
	grpcServer *grpc.Server
	stopOnce   sync.Once
}

var GDSCServer *DSCServer

func (dscServer *DSCServer) Run() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", dscServer.port))
	if err != nil {
		return err
	}
	pb.RegisterRpcGameDscServer(dscServer.grpcServer, &gamedsc.RPCGameDscServer{})
	pb.RegisterRpcDsaReportServer(dscServer.grpcServer, &dsareport.Server{})
	logrus.WithField("port", dscServer.port).Info("dsc grpc serve")
	return dscServer.grpcServer.Serve(lis)
}

func (dscServer *DSCServer) Stop() {
	dscServer.stopOnce.Do(func() {
		if dscServer.grpcServer != nil {
			dscServer.grpcServer.GracefulStop()
		}
	})
}

func NewDSCServer(port int) (dscServer *DSCServer, err error) {
	dscServer = &DSCServer{
		port:       port,
		grpcServer: grpc.NewServer(),
	}
	GDSCServer = dscServer
	return
}

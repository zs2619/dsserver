package dsaclient

import (
	"context"
	"dsservices/pb"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	mu    sync.Mutex
	conns = make(map[string]*grpc.ClientConn)
)

func getConn(addr string) (*grpc.ClientConn, error) {
	mu.Lock()
	defer mu.Unlock()
	if c, ok := conns[addr]; ok {
		return c, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, err
	}
	conns[addr] = conn
	return conn, nil
}

func CreateDS(ctx context.Context, grpcAddr string, req *pb.DsaCreateDSReq) (*pb.DsaCreateDSResp, error) {
	conn, err := getConn(grpcAddr)
	if err != nil {
		logrus.WithError(err).WithField("addr", grpcAddr).Error("dsaclient dial")
		return nil, err
	}
	client := pb.NewRpcDsaAgentClient(conn)
	return client.CreateDS(ctx, req)
}

func RemoveDS(ctx context.Context, grpcAddr string, req *pb.DsaRemoveDSReq) (*pb.DsaRemoveDSResp, error) {
	conn, err := getConn(grpcAddr)
	if err != nil {
		return nil, err
	}
	client := pb.NewRpcDsaAgentClient(conn)
	return client.RemoveDS(ctx, req)
}

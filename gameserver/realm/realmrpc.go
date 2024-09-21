package realm

import (
	"dsservices/pb"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var dscGameClient pb.RpcGameDscClient

func DSCGameGRPCInit(serverAddr string) error {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())
	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Info("game server error")
	}
	logrus.WithFields(
		logrus.Fields{
			"address": serverAddr,
		}).Info("connect dsc ok")

	dscGameClient = pb.NewRpcGameDscClient(conn)
	return nil
}

func GetRpcClient() pb.RpcGameDscClient {
	return dscGameClient
}

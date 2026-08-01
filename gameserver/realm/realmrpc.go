package realm

import (
	"dsservices/pb"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var dscGameClient pb.RpcGameDscClient

func DSCGameGRPCInit(serverAddr string) error {
	conn, err := grpc.Dial(serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		logrus.WithError(err).Error("dial dsc error")
		return err
	}
	logrus.WithField("address", serverAddr).Info("connect dsc ok")
	dscGameClient = pb.NewRpcGameDscClient(conn)
	return nil
}

func GetRpcClient() pb.RpcGameDscClient {
	return dscGameClient
}

package realm

import (
	"context"
	"dsservices/pb"

	"github.com/sirupsen/logrus"
)

func CreateRealm(realmCfgID string) error {
	_, err := GetRpcClient().CreateDS(context.TODO(), &pb.RpcCreateDSReq{})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Info("CreateRealm")
	}
	return nil
}

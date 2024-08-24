package realm

import (
	"context"
	"dsservices/pb"

	"github.com/sirupsen/logrus"
)

func CreateRealm(realmCfgID string) error {
	_, err := GetRpcClient().CreateRealm(context.TODO(), &pb.RpcCreateRealmReq{})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Info("CreateRealm")
	}
	return nil
}

package dsa

import (
	"dsservices/pb"
	"context"
	"io"
	"time"

	"github.com/sirupsen/logrus"
)

var CreateRealmResultChan chan *pb.RpcCreateRealmResult = make(chan *pb.RpcCreateRealmResult, 2048)

func RunWaitCreateRealm(client pb.DsaDscARealmClient) {
	var stream pb.DsaDscARealm_WaitCreateRealmClient
	var err error
	for {
		stream, err = client.WaitCreateRealm(context.Background())
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Info("client.WaitCreateRealm error")
			time.Sleep(time.Second * 3)
		} else {
			logrus.WithFields(logrus.Fields{}).Info("client.WaitCreateRealm ok")
			break
		}
	}

	go func() {
		for resp := range CreateRealmResultChan {
			if err := stream.Send(resp); err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Info("client.WaitCreateRealm send error")
			}
		}
	}()

	for {
		info, err := stream.Recv()
		if err == io.EOF {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Info("client.WaitCreateRealm recv finish")
			return
		}
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Info("client.WaitCreateRealm recv error")
			time.Sleep(time.Second)
			continue
		}

		///创建DS
		_, err = NewDS(info.DsID, info.RealmCfgID, info.TeamID)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Info("proc.StartProc  error")
		}
	}
}

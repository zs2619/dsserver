package realm

import (
	"dsservices/pb"
	"context"
	"io"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var realmClient pb.GameDscRealmClient

func RealmGRPCInit(serverAddr string) error {
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

	realmClient = pb.NewGameDscRealmClient(conn)
	go runNotifyJoinRealm()
	return nil
}

func GetRpcClient() pb.GameDscRealmClient {
	return realmClient
}

var NotifyJoinRealmChan chan *pb.NotifyJoinRealmReq = make(chan *pb.NotifyJoinRealmReq, 2048)

func runNotifyJoinRealm() error {
	var stream pb.GameDscRealm_NotifyJoinRealmClient
	wg := sync.WaitGroup{}
	wg.Add(2)
	for {
		var err error
		stream, err = realmClient.NotifyJoinRealm(context.Background())
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Info("client.NotifyJoinRealm error")
			time.Sleep(time.Second * 3)
		} else {
			logrus.WithFields(logrus.Fields{}).Info("client.NotifyJoinRealm ok")
			break
		}
	}
	go func() {
		for req := range NotifyJoinRealmChan {
			err := stream.Send(req)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Info("client.NotifyJoinRealm recv finish")
				break
			}
		}
		wg.Done()
	}()

	go func() {
		for {
			info, err := stream.Recv()
			if err == io.EOF {
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Info("client.NotifyJoinRealm recv finish")
				break
			}
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Info("client.NotifyJoinRealm recv error")
				time.Sleep(time.Second)
				continue
			}

			//通知客户端加入副本
			notifyPlayerJoinRealm(info.TeamID, info.DsAddr, info.RealmCfgID)
		}
		wg.Done()
	}()
	wg.Wait()
	logrus.WithFields(logrus.Fields{}).Info("runNotifyJoinRealm quit")
	return nil
}

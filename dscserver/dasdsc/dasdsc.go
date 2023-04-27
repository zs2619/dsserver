package dasdsc

import (
	"dsservices/pb"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type RPCDasDscServer struct {
	pb.UnimplementedDsaDscARealmServer
}

var CreateRealmChan chan *pb.RpcCreateRealmInfo = make(chan *pb.RpcCreateRealmInfo, 2048)
var CreateRealmResult chan *pb.RpcCreateRealmResult = make(chan *pb.RpcCreateRealmResult, 2048)

func (rpcServer *RPCDasDscServer) WaitCreateRealm(realm pb.DsaDscARealm_WaitCreateRealmServer) error {
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		for {
			data, err := realm.Recv()
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
					"data":  data,
				}).Info("WaitCreateRealm error")
				break
			}
			CreateRealmResult <- data
		}
		wg.Done()
		CreateRealmChan <- nil
	}()

	go func() {
		for info := range CreateRealmChan {
			if info == nil {
				break
			}
			realm.Send(info)
			time.Sleep(time.Microsecond * 100)
		}
		wg.Done()
	}()

	wg.Wait()

	logrus.WithFields(logrus.Fields{}).Info("WaitCreateRealm quit")
	return nil
}
func (rpcServer *RPCDasDscServer) JoinRealm(realm pb.DsaDscARealm_WaitJoinRealmServer) error {
	return nil
}

func CreateRealm(dsID, teamID, realmCfgID string) error {
	info := &pb.RpcCreateRealmInfo{DsID: dsID, TeamID: teamID, RealmCfgID: realmCfgID}
	CreateRealmChan <- info
	return nil
}

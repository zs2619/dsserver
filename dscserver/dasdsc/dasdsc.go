package dasdsc

import (
	"dsservices/pb"

	"github.com/sirupsen/logrus"
)

type RPCDasDscServer struct {
	pb.UnimplementedDsaDscARealmServer
}

func (rpcServer *RPCDasDscServer) StreamService(realm pb.DsaDscARealm_StreamServiceServer) error {
	// wg := sync.WaitGroup{}
	// wg.Add(2)
	// go func() {
	// 	for {
	// 		data, err := realm.Recv()
	// 		if err != nil {
	// 			logrus.WithFields(logrus.Fields{
	// 				"error": err,
	// 				"data":  data,
	// 			}).Info("WaitCreateRealm error")
	// 			break
	// 		}
	// 	}
	// 	wg.Done()
	// }()

	// go func() {
	// 	for info := range CreateRealmChan {
	// 		if info == nil {
	// 			break
	// 		}
	// 		realm.Send(info)
	// 		time.Sleep(time.Microsecond * 100)
	// 	}
	// 	wg.Done()
	// }()

	// wg.Wait()

	logrus.WithFields(logrus.Fields{}).Info("WaitCreateRealm quit")
	return nil
}

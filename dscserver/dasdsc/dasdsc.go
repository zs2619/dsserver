package dasdsc

import (
	"dsservices/common"
	"dsservices/dscserver/dsamgr"
	"dsservices/pb"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

type RPCDasDscServer struct {
	pb.UnimplementedDsaDscARealmServer
}

func (rpcServer *RPCDasDscServer) StreamService(stream pb.DsaDscARealm_StreamServiceServer) error {
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return fmt.Errorf("peer.FromContext error")
	}
	if ok && len(md.Get(common.MD_KEY_AGENTID)) == 0 {
		return fmt.Errorf("metadata name error")
	}
	agentName := md.Get(common.MD_KEY_AGENTID)[0]

	peer, ok := peer.FromContext(stream.Context())
	if !ok {
		return fmt.Errorf("metadata peer error")
	}

	logrus.WithFields(logrus.Fields{
		"agentName": agentName,
		"addr":      peer.Addr.String(),
	}).Info("StreamService init ok")

	dsa := dsamgr.GDSAMgr.Get(agentName)
	if dsa == nil {
		// 创建
		dsa = dsamgr.NewDSAInfo(agentName, peer.Addr.String())
		dsamgr.GDSAMgr.Add(dsa)
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		for {
			data, err := stream.Recv()
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
					"data":  data,
				}).Info("WaitCreateRealm error")
				break
			}
		}
		wg.Done()
	}()

	go func() {
		for info := range dsa.StreamServerEventChan {
			if info == nil {
				break
			}
			stream.Send(info)
			time.Sleep(time.Microsecond * 100)
		}
		wg.Done()
	}()

	wg.Wait()

	logrus.WithFields(logrus.Fields{"agentName": agentName}).Info("StreamService end")
	return nil
}

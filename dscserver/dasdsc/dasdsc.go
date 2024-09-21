package dasdsc

import (
	"dsservices/common"
	"dsservices/dscserver/dsamgr"
	"dsservices/pb"
	"fmt"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

type RPCDasDscServer struct {
	pb.UnimplementedStreamDscDSAServer
}

func (rpcServer *RPCDasDscServer) StreamService(stream pb.StreamDscDSA_StreamServiceServer) error {
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
		// create new dsainfo
		dsa = dsamgr.NewDSAInfo(agentName, peer.Addr.String(), stream)
		dsamgr.GDSAMgr.Add(dsa)
		dsa.Run()
	} else {
		//TODO remove old and add new
	}

	logrus.WithFields(logrus.Fields{"agentName": agentName}).Info("StreamService end")
	return nil
}

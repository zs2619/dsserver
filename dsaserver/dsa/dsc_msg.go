package dsa

import (
	"dsservices/pb"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

func RouterMap(agent *DSANatsConn, pmsg proto.Message) (err error) {
	logrus.WithFields(logrus.Fields{"pmsg": pmsg}).Info("RouterMap")

	switch msg := pmsg.(type) {
	case *pb.StreamCreateDS:
		err = CreateDS(agent, msg)
	case *pb.StreamRemoveDS:
		err = RemoveDS(agent, msg)
	default:
		{
			logrus.WithError(errors.WithStack(err)).WithFields(logrus.Fields{"m": pmsg}).Error("routerMap default")
		}
	}
	return
}
func CreateDS(agent *DSANatsConn, msg *pb.StreamCreateDS) (err error) {
	logrus.WithFields(logrus.Fields{"m": msg}).Info("CreateDS")
	return nil
}

func RemoveDS(agent *DSANatsConn, msg *pb.StreamRemoveDS) (err error) {
	logrus.WithFields(logrus.Fields{"m": msg}).Info("RemoveDS")
	return nil
}

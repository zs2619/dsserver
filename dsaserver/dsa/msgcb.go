package dsa

import (
	"dsservices/kissnet"
	"dsservices/pb"
	"encoding/binary"
	"fmt"

	"github.com/sirupsen/logrus"
)

type DSAServerCB struct {
}

var GDSAServerCB DSAServerCB

func (dsaCB *DSAServerCB) OnDisConectCB(conn kissnet.IConnection) error {
	logrus.WithFields(logrus.Fields{}).Info("OnDisConectCB")
	return nil
}

func (dsaCB *DSAServerCB) OnConectCB(conn kissnet.IConnection) error {
	logrus.WithFields(logrus.Fields{}).Info("OnConectCB")
	return nil
}

func (dsaCB *DSAServerCB) OnMsgCB(conn kissnet.IConnection, msg []byte) error {
	if msg == nil {
		return nil
	}
	if len(msg) < 2 {
		return fmt.Errorf("dsa msg len error")
	}

	msgID := pb.DS2DSA_MsgID_MsgID(binary.LittleEndian.Uint16(msg))
	err := ProcDSMsg(conn, msgID, msg[2:])
	if err != nil {
		return nil
	}

	return nil
}
func ProcDSMsg(conn kissnet.IConnection, msgID pb.DS2DSA_MsgID_MsgID, msg []byte) error {
	if msgID == pb.DS2DSA_MsgID_DSCreateOK {
		err := dsCreateOK(conn, msg)
		return err
	} else {
		f, ok := msgMap[msgID]
		if !ok {
			return fmt.Errorf("(%d) msgID nil", msgID)
		}
		u := GDSInfoMgr.GetDSByConn(conn)
		if u == nil {
			return fmt.Errorf("(%d) ProcMsg ds nil", msgID)
		}
		return f(u, msg)
	}
}

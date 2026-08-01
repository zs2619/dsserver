package gameplayer

import (
	"encoding/binary"
	"fmt"

	"dsservices/common"
	"dsservices/kissnet"
	"dsservices/pb"

	"github.com/sirupsen/logrus"
)

type GameServerCB struct {
}

var GGamePlayerCB GameServerCB

func (gsCB *GameServerCB) OnDisConectCB(conn kissnet.IConnection) error {
	logrus.WithFields(logrus.Fields{}).Info("OnDisConectCB")
	inRealm := GPlayerMgr.MarkDisconnected(conn, common.GetNowTime().Unix())
	if !inRealm {
		// not in a realm: drop immediately (team matching handled by timeout cron)
		GPlayerMgr.DelPlayer(conn)
	}
	return nil
}

func (gsCB *GameServerCB) OnConectCB(conn kissnet.IConnection) error {
	logrus.WithFields(logrus.Fields{}).Info("OnConectCB")
	return nil
}

func (gsCB *GameServerCB) OnMsgCB(conn kissnet.IConnection, msg []byte) error {
	if msg == nil {
		GPlayerMgr.DelPlayer(conn)
		return nil
	}
	if len(msg) < 2 {
		GPlayerMgr.DelPlayer(conn)
		return fmt.Errorf("game msg len error")
	}
	msgID := pb.C2S_MsgID_MsgID(binary.LittleEndian.Uint16(msg))
	err := ProcGameMsg(conn, msgID, msg[2:])
	if err != nil {
		return nil
	}
	return nil
}

func ProcGameMsg(conn kissnet.IConnection, msgID pb.C2S_MsgID_MsgID, msg []byte) error {
	if msgID == pb.C2S_MsgID_LocalLogin {
		return LocalLogin(conn, msgID, msg)
	} else if msgID == pb.C2S_MsgID_AuthLogin {
		return AuthLogin(conn, msgID, msg)
	} else {
		f, ok := msgMap[msgID]
		if !ok {
			return fmt.Errorf("(%d) msgID nil", msgID)
		}
		u := GPlayerMgr.GetPlayerByClientID(conn)
		if u == nil {
			return fmt.Errorf("(%d) ProcMsg player nil", msgID)
		}
		return f(u, msg)
	}

}

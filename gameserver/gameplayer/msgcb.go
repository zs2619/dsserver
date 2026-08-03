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

func (gsCB *GameServerCB) OnDisconnectCB(conn kissnet.Connection) error {
	logrus.WithFields(logrus.Fields{}).Info("OnDisconnectCB")
	inRealm := GPlayerMgr.MarkDisconnected(conn, common.GetNowTime().Unix())
	if !inRealm {
		// not in a realm: drop immediately (team matching handled by timeout cron)
		GPlayerMgr.DelPlayer(conn)
	}
	return nil
}

func (gsCB *GameServerCB) OnConnectCB(conn kissnet.Connection) error {
	logrus.WithFields(logrus.Fields{}).Info("OnConnectCB")
	return nil
}

func (gsCB *GameServerCB) OnMsgCB(conn kissnet.Connection, msg []byte) error {
	if msg == nil {
		GPlayerMgr.DelPlayer(conn)
		return nil
	}
	if len(msg) < 2 {
		GPlayerMgr.DelPlayer(conn)
		conn.Close()
		return fmt.Errorf("game msg len error")
	}
	msgID := pb.C2S_MsgID_MsgID(binary.LittleEndian.Uint16(msg))
	err := ProcGameMsg(conn, msgID, msg[2:])
	if err == nil {
		return nil
	}
	logrus.WithError(err).WithField("msgID", msgID).Warn("ProcGameMsg")
	// 登录失败已发 LoginFail，连接保留便于重试；其它协议错误断连，避免客户端空等。
	if msgID != pb.C2S_MsgID_LocalLogin && msgID != pb.C2S_MsgID_AuthLogin {
		conn.Close()
	}
	return err
}

func ProcGameMsg(conn kissnet.Connection, msgID pb.C2S_MsgID_MsgID, msg []byte) error {
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

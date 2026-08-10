package gameplayer

import (
	"bytes"
	"dsservices/common"
	"github.com/zs2619/kissnet-go"
	"dsservices/pb"
	"encoding/binary"

	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

// LoginFAILResp.errCode（客户端约定，proto 未枚举）
const (
	LoginErrInternal   int64 = 1
	LoginErrBadRequest int64 = 2
)

func sendToConn(conn kissnet.Connection, msgID pb.S2C_MsgID_MsgID, body []byte) error {
	if conn == nil {
		return nil
	}
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, uint16(msgID)); err != nil {
		return err
	}
	buf.Write(body)
	return conn.SendMsg(buf)
}

func sendLoginFail(conn kissnet.Connection, errCode int64) {
	m, err := proto.Marshal(&pb.LoginFAILResp{ErrCode: errCode})
	if err != nil {
		logrus.WithError(err).Error("marshal LoginFAILResp")
		return
	}
	if err := sendToConn(conn, pb.S2C_MsgID_LoginFail, m); err != nil {
		logrus.WithError(err).Warn("send LoginFail")
	}
}

func (player *GamePlayer) SendLocalLoginResp() error {
	logrus.Debug("SendLoginMsg")

	pbpm := player.GetPlayerModelPB()
	pbMsg := &pb.LoginOKResp{
		ServerTime:  common.GetNowTime().UnixMilli(),
		PlayerModel: pbpm,
	}
	m, err := proto.Marshal(pbMsg)
	if err != nil {
		return err
	}

	return player.SendMsg(pb.S2C_MsgID_LoginOK, m)
}

func (player *GamePlayer) SendJoinRealmOKResp(teamID, realmCfgID, DsAddr string) error {
	logrus.Debug("SendJoinRealmOKResp")

	pbMsg := &pb.JoinRealmOKResp{RealmConfigID: realmCfgID, DsAddr: DsAddr, TeamID: teamID}
	m, err := proto.Marshal(pbMsg)
	if err != nil {
		return err
	}
	return player.SendMsg(pb.S2C_MsgID_JoinRealmOK, m)
}

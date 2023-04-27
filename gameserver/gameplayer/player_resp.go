package gameplayer

import (
	"dsservices/common"
	"dsservices/pb"

	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

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

	player.SendMsg(pb.S2C_MsgID_LoginOK, m)
	return nil
}

func (player *GamePlayer) SendJoinRealmOKResp(teamID, realmCfgID, DsAddr string) error {
	logrus.Debug("SendJoinRealmOKResp")

	pbMsg := &pb.JoinRealmOKResp{RealmConfigID: realmCfgID, DsAddr: DsAddr, TeamID: teamID}
	m, _ := proto.Marshal(pbMsg)

	player.SendMsg(pb.S2C_MsgID_JoinRealmOK, m)

	return nil
}

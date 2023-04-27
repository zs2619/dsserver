package gameplayer

import (
	"dsservices/gameserver/team"
	"dsservices/pb"
	"fmt"

	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

type FuncMsg func(*GamePlayer, []byte) error

var (
	msgMap map[pb.C2S_MsgID_MsgID]FuncMsg = make(map[pb.C2S_MsgID_MsgID]FuncMsg)
)

func init() {
	RegisterMsgMapFunc()
}

func RegisterMsgMapFunc() {
	msgMap[pb.C2S_MsgID_Ping] = pingReq
	msgMap[pb.C2S_MsgID_QuickJoinRealm] = quickJoinRealmReq
	msgMap[pb.C2S_MsgID_StartRealm] = startRealmReq
	msgMap[pb.C2S_MsgID_QuitRealmTeam] = quitRealmTeam
}

func pingReq(player *GamePlayer, msg []byte) error {
	return nil
}

func quickJoinRealmReq(player *GamePlayer, msg []byte) error {
	logrus.Debug("quickJoinRealm")

	joinRealm := &pb.QuickJoinRealmReq{}
	err := proto.Unmarshal(msg, joinRealm)
	if err != nil {
		return err
	}
	if len(player.TeamID) != 0 {
		return fmt.Errorf("len(player.TeamID) != 0")
	}

	_, err = team.QuickJoinRealmTeam(player.UserID, joinRealm.RealmConfigID)
	if err != nil {
		return err
	}

	return nil
}

func startRealmReq(player *GamePlayer, msg []byte) error {
	logrus.Debug("startRealmReq")
	startRealm := &pb.StartRealmReq{}
	err := proto.Unmarshal(msg, startRealm)
	if err != nil {
		return err
	}
	if len(player.TeamID) == 0 {
		return fmt.Errorf("len(player.TeamID) == 0")
	}
	err = team.StartRealm(player.UserID, player.TeamID)
	if err != nil {
		return err
	}
	return nil
}

func quitRealmTeam(player *GamePlayer, msg []byte) error {
	logrus.Debug("cancelRealmTeam")
	req := &pb.QuitRealmTeamReq{}
	err := proto.Unmarshal(msg, req)
	if err != nil {
		return err
	}
	if len(player.TeamID) == 0 {
		return fmt.Errorf("len(player.TeamID) == 0")
	}
	_, err = team.QuitRealmTeam(player.UserID, player.TeamID)
	if err != nil {
		return err
	}
	return nil
}

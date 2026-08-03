package gameplayer

import (
	"dsservices/gameserver/realm"
	"dsservices/pb"

	"google.golang.org/protobuf/proto"
)

func init() {
	realm.LocalPlayerExists = func(userID string) bool {
		return GPlayerMgr.GetPlayerByUserID(userID) != nil
	}
	realm.PushToLocalPlayers = func(userIDs []string, realmID, teamID, realmCfgID, dsAddr string) {
		// 拷贝切片，避免调用方并发改底层数组
		uids := append([]string{}, userIDs...)
		schedule(func() {
			for _, uid := range uids {
				p := GPlayerMgr.GetPlayerByUserID(uid)
				if p == nil {
					continue
				}
				p.TeamID = teamID
				p.RealmID = realmID
				_ = p.SendJoinRealmOKResp(teamID, realmCfgID, dsAddr)
			}
		})
	}
	realm.PushRealmFailToLocalPlayers = func(userIDs []string, realmID string) {
		uids := append([]string{}, userIDs...)
		schedule(func() {
			for _, uid := range uids {
				p := GPlayerMgr.GetPlayerByUserID(uid)
				if p == nil {
					continue
				}
				if p.RealmID == realmID {
					p.RealmID = ""
				}
				p.TeamID = ""
				msg, _ := proto.Marshal(&pb.JoinRealmFailResp{})
				_ = p.SendMsg(pb.S2C_MsgID_JoinRealmFail, msg)
			}
		})
	}
}

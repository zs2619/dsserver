package gameplayer

import (
	"dsservices/gameserver/realm"
	"dsservices/gameserver/team"
)

func init() {
	go NotifyPlayerJoinRealm()
}

func NotifyPlayerJoinRealm() {
	for playerJoinRealm := range realm.PlayerJoinRealmChan {
		userList, err := team.GetTeamUserList(playerJoinRealm.TeamID)
		if err != nil {
			continue
		}
		for _, v := range userList {
			player := GPlayerMgr.GetPlayerByUserID(v)
			if player != nil {
				//TODo:暂时去掉
				player.SendJoinRealmOKResp(playerJoinRealm.TeamID[7:], playerJoinRealm.RealmCfgID, playerJoinRealm.DSAddr)
			}
		}
	}
}

package gameplayer

import (
	"dsservices/gameserver/playermodel"
	"dsservices/gameserver/user"
	"dsservices/kissnet"
	"dsservices/pb"

	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/protobuf/proto"
)

func LocalLogin(conn kissnet.IConnection, msgID pb.C2S_MsgID_MsgID, msg []byte) error {
	loginMsg := &pb.LocalLoginReq{}
	err := proto.Unmarshal(msg, loginMsg)
	if err != nil {
		return err
	}
	err = login(loginMsg.UserID, conn)
	return err
}

func AuthLogin(conn kissnet.IConnection, msgID pb.C2S_MsgID_MsgID, msg []byte) error {
	//TODO:验证
	loginMsg := &pb.AuthLoginReq{}
	err := proto.Unmarshal(msg, loginMsg)
	if err != nil {
		return err
	}
	userID := loginMsg.SessionID

	err = login(userID, conn)
	return err
}

func login(userID string, conn kissnet.IConnection) error {
	player := GPlayerMgr.GetPlayerByUserID(userID)
	if player != nil {
		// 踢掉之前连接
		GPlayerMgr.DelPlayer(player.Conn)
	}

	userRet, err := getUser(userID, user.PlatFormLocal, "ip")
	if err != nil {
		return err
	}

	pm, err := getLoginData(userRet)
	if err != nil {
		return err
	}

	newPlayer := NewPlayer(pm, conn)

	err = GPlayerMgr.AddPlayer(newPlayer, conn)
	if err != nil {
		return err
	}

	err = newPlayer.SendLocalLoginResp()

	return err
}

func getUser(userID, platform, ip string) (userRet *user.UserType, err error) {
	userRet, err = user.GetUser(userID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err = playermodel.CreateUserPlayer(userID, platform, ip)
			if err != nil {
				return
			}
			userRet, err = user.GetUser(userID)
			if err != nil {
				return
			}
		}
	}
	return
}

func getLoginData(userRet *user.UserType) (pm *playermodel.PlayerModel, err error) {
	pm, err = playermodel.LoadPlayerModel(userRet.PlayerID.Hex())
	return
}

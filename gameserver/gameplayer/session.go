package gameplayer

import (
	"fmt"

	"dsservices/gameserver/playermodel"
	"dsservices/gameserver/realm"
	"dsservices/gameserver/user"
	"github.com/zs2619/kissnet-go"
	"dsservices/pb"

	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/protobuf/proto"
)

func LocalLogin(conn kissnet.Connection, msgID pb.C2S_MsgID_MsgID, msg []byte) error {
	loginMsg := &pb.LocalLoginReq{}
	if err := proto.Unmarshal(msg, loginMsg); err != nil {
		sendLoginFail(conn, LoginErrBadRequest)
		return err
	}
	if err := login(loginMsg.UserID, conn); err != nil {
		sendLoginFail(conn, loginFailCode(err))
		return err
	}
	return nil
}

func AuthLogin(conn kissnet.Connection, msgID pb.C2S_MsgID_MsgID, msg []byte) error {
	//TODO:验证
	loginMsg := &pb.AuthLoginReq{}
	if err := proto.Unmarshal(msg, loginMsg); err != nil {
		sendLoginFail(conn, LoginErrBadRequest)
		return err
	}
	if err := login(loginMsg.SessionID, conn); err != nil {
		sendLoginFail(conn, loginFailCode(err))
		return err
	}
	return nil
}

func loginFailCode(err error) int64 {
	if err == errLoginBadRequest {
		return LoginErrBadRequest
	}
	return LoginErrInternal
}

var errLoginBadRequest = fmt.Errorf("login bad request")

func login(userID string, conn kissnet.Connection) error {
	if userID == "" {
		return errLoginBadRequest
	}
	player, oldConn := GPlayerMgr.RebindConn(userID, conn)
	if player != nil {
		if oldConn != nil {
			oldConn.Close()
		}
	} else {
		userRet, err := getUser(userID, user.PlatFormLocal, kissnet.PeerIP(conn))
		if err != nil {
			return err
		}
		pm, err := getLoginData(userRet)
		if err != nil {
			return err
		}
		newPlayer := NewPlayer(pm, conn)
		if err := GPlayerMgr.AddPlayer(newPlayer, conn); err != nil {
			return err
		}
		player = newPlayer
	}

	err := player.SendLocalLoginResp()
	if err != nil {
		return err
	}

	realm.TryPushRealmOnLogin(userID, func(realmID, teamID, realmCfgID, dsAddr string) {
		player.RealmID = realmID
		player.TeamID = teamID
		_ = player.SendJoinRealmOKResp(teamID, realmCfgID, dsAddr)
	})
	return nil
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

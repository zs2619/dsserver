package gameplayer

import (
	"dsservices/gameserver/playermodel"
	"dsservices/kissnet"
	"dsservices/pb"
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/sirupsen/logrus"
)

type GamePlayer struct {
	UserID      string
	TeamID      string
	Conn        kissnet.IConnection
	PlayerModel *playermodel.PlayerModel
}

func NewPlayer(playerModel *playermodel.PlayerModel, conn kissnet.IConnection) *GamePlayer {
	player := &GamePlayer{
		UserID:      playerModel.UserID,
		Conn:        conn,
		PlayerModel: playerModel,
	}
	player.Init()
	return player
}

func (player *GamePlayer) Init() {
	logrus.WithFields(logrus.Fields{"userID": player.UserID}).Info("GamePlayer Init")
}

func (player *GamePlayer) Dispose() {
	logrus.WithFields(logrus.Fields{"userID": player.UserID}).Info("GamePlayer Dispose")
}

func (player *GamePlayer) SendMsg(msgID pb.S2C_MsgID_MsgID, msg []byte) error {
	sendMsg := new(bytes.Buffer)
	binary.Write(sendMsg, binary.LittleEndian, uint16(msgID))
	sendMsg.Write(msg)
	if player.Conn == nil {
		return fmt.Errorf("this.Conn == nil")
	}
	player.Conn.SendMsg(sendMsg)
	return nil
}
func (player *GamePlayer) GetPlayerModelPB() *pb.PlayerModel {
	retPB := &pb.PlayerModel{
		UserID:     player.PlayerModel.UserID,
		PlayerID:   player.PlayerModel.PlayerID.Hex(),
		Level:      player.PlayerModel.Level,
		Exp:        player.PlayerModel.Exp,
		Name:       player.PlayerModel.PlayerName,
		CreateTime: player.PlayerModel.CreateTime,
	}
	return retPB
}

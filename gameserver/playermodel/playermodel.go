package playermodel

import (
	"dsservices/common"
	"dsservices/gameserver/user"
	"dsservices/store"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PlayerModel struct {
	PlayerID   primitive.ObjectID `json:"playerID" bson:"_id"`
	UserID     string             `json:"userID" bson:"userID"`
	PlayerName string             `json:"name" bson:"name"`
	CreateTime int64              `json:"createTime" bson:"createTime"`
	IP         string             `json:"ip" bson:"ip"`
	Level      int64              `json:"level" bson:"level"`
	Exp        int64              `json:"exp" bson:"exp"`
}

func CreatePlayerModel(userID string, playerObj primitive.ObjectID, ip string) (*PlayerModel, error) {
	currTime := common.GetNowTime().UnixMilli()
	pm := &PlayerModel{
		PlayerID:   playerObj,
		UserID:     userID,
		PlayerName: "",
		CreateTime: currTime,
		IP:         ip,
		Level:      1,
		Exp:        0,
	}
	_, err := store.GetColl(store.PlayerColl).InsertOne(context.Background(), pm)
	if err != nil {
		return nil, err
	}
	return pm, nil

}

func CreateUserPlayer(userID, platform, ip string) error {
	playerID := primitive.NewObjectID()
	user, err := user.CreateUser(userID, platform, playerID)
	if err != nil {
		return err
	}
	_, err = CreatePlayerModel(userID, user.PlayerID, ip)
	if err != nil {
		return err
	}
	return nil
}
func LoadPlayerModel(playerID string) (*PlayerModel, error) {
	uidObj, err := primitive.ObjectIDFromHex(playerID)
	if err != nil {
		return nil, err
	}
	var pm PlayerModel
	filter := bson.D{{"_id", uidObj}}
	err = store.GetColl(store.PlayerColl).FindOne(context.TODO(), filter).Decode(&pm)
	if err != nil {
		return nil, err
	}
	return &pm, nil
}

package user

import (
	"dsservices/store"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	PlatFormLocal = "local"
	GMLevelLocal  = 0
)

type UserType struct {
	UserID   string             `json:"userID" bson:"userID"`
	Platform string             `json:"platform" bson:"platform"`
	PlayerID primitive.ObjectID `json:"playerID" bson:"playerID"`
	GmLevel  int                `json:"gmLevel" bson:"gmLevel"`
}

func CreateUser(userID, platform string, uidObj primitive.ObjectID) (*UserType, error) {
	var gmLevel int
	if platform == PlatFormLocal {
		gmLevel = 1
	} else {
		gmLevel = 0
	}

	user := &UserType{
		UserID:   userID,
		Platform: platform,
		PlayerID: uidObj,
		GmLevel:  gmLevel,
	}
	_, err := store.GetColl(store.UserColl).InsertOne(context.Background(), user)
	if err != nil {
		return nil, err
	}
	return user, err
}

func GetUser(userID string) (*UserType, error) {
	var result UserType
	err := store.GetColl(store.UserColl).FindOne(context.Background(), bson.D{{"userID", userID}}).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func DelUser(userID string) error {
	filter := bson.D{{"userID", userID}}
	_, err := store.GetColl(store.UserColl).DeleteOne(context.TODO(), filter)
	return err
}

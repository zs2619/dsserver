package store

import (
	"context"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var MongoDB *mongo.Database

const (
	UserColl   = "user"
	PlayerColl = "player"
)

func mongoConnect(mongoDBURI, mongoDBName string) (*mongo.Database, error) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().
		ApplyURI(mongoDBURI).
		SetMinPoolSize(100)

	logrus.WithFields(logrus.Fields{
		"URI": mongoDBURI,
	}).Info("mongoConnect")

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"dbName": mongoDBName,
	}).Info("mongoConnect")

	MongoDB := client.Database(mongoDBName)
	return MongoDB, err

}
func MongoInit() error {
	mongoDBURI := os.Getenv("DS_MONGOURI")
	mongoDBName := os.Getenv("DS_MONGODBNAME")

	var err error
	MongoDB, err = mongoConnect(mongoDBURI, mongoDBName)
	if err != nil {
		return err
	}
	err = createIndex()
	return err
}
func MongoFinish() error {
	logrus.WithFields(logrus.Fields{}).Info("MongoFinish")
	if MongoDB != nil {
		return MongoDB.Client().Disconnect(context.TODO())
	}
	return nil
}
func GetColl(name string) *mongo.Collection {
	return MongoDB.Collection(name)
}

func createIndex() error {
	err := createUserIndex()
	if err != nil {
		return err
	}
	err = createPlayerIndex()
	if err != nil {
		return err
	}
	return nil
}

func createUserIndex() error {
	indexOptions := options.Index()
	indexOptions.SetUnique(true)
	models := []mongo.IndexModel{
		{
			Keys:    bson.D{{"userID", 1}},
			Options: indexOptions,
		},
	}
	_, err := GetColl(UserColl).Indexes().CreateMany(context.TODO(), models)
	return err
}

func createPlayerIndex() error {
	indexOptions := options.Index()
	indexOptions.SetUnique(true)
	models := []mongo.IndexModel{
		{
			Keys:    bson.D{{"userID", 1}},
			Options: indexOptions,
		},
	}
	_, err := GetColl(PlayerColl).Indexes().CreateMany(context.TODO(), models)
	return err
}

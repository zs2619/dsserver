package store

import (
	"context"
	"os"
	"strings"

	"github.com/redis/go-redis/v9"

	"github.com/sirupsen/logrus"
)

var RedisClient redis.UniversalClient

func RedisInit() error {
	redisURI := os.Getenv("DS_REDISURI")

	var err error
	redisClusterAddrs := strings.Split(redisURI, ",")
	RedisClient = redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs: redisClusterAddrs,
	})
	logrus.WithFields(logrus.Fields{
		"URI": redisURI,
	}).Info("RedisInit")

	_, err = RedisClient.Ping(context.TODO()).Result()
	if err != nil {
		return err
	}
	return nil
}

func RedisFinish() error {
	logrus.WithFields(logrus.Fields{}).Info("RedisFinish")
	return RedisClient.Close()
}

func hashTag(uid string) string {
	return "{" + uid + "}"
}

func FlushAll() {
	RedisClient.FlushAll(context.TODO())
	clusterClient, ok := RedisClient.(*redis.ClusterClient)
	if ok {
		clusterClient.ForEachMaster(context.TODO(), func(ctx context.Context, client *redis.Client) error {
			client.FlushAll(context.TODO())
			return nil
		})
	}
}

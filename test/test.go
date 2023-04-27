package test

import (
	"dsservices/config"
	"dsservices/store"
	"context"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

type BaseTestSuite struct {
	suite.Suite
}

func init() {
	var once sync.Once
	onceInit := func() {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		_, err := config.LoadGameConfig("assets/config.json")
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("config.LoadGameConfig")
		}
		logLevel := os.Getenv("DS_LOGLEVEL")
		if len(logLevel) == 0 || logLevel == "debug" {
			logrus.SetLevel(logrus.DebugLevel)
		}
	}
	once.Do(onceInit)
}

func (tb *BaseTestSuite) SetupSuite() {
	err := store.InitStore()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Info("store.InitStore error")
		return
	}
}

func (tb *BaseTestSuite) TearDownSuite() {
	logrus.Info("Tearing down")
	store.MongoDB.Drop(context.TODO())
	store.FlushAll()
	store.FinishStore()
}

type NoDBTestSuite struct {
	suite.Suite
}

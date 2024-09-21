package common

import (
	"dsservices/config"

	"github.com/sirupsen/logrus"
)

func Init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.Info("GitVersion:", GitVersion)
	_, err := config.LoadGameConfig("assets/config.json")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Info("config.LoadGameConfig error")
		return
	}
	logrus.Info("config.LoadGameConfig OK")

	InitCron()

}

func Finish() {
	FinishCron()
}

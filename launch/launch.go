package launch

import (
	"dsservices/common"
	"dsservices/config"

	"github.com/sirupsen/logrus"
)

func Init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.Info("GitVersion:", common.GitVersion)
	_, err := config.LoadGameConfig("assets/config.json")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Info("config.LoadGameConfig error")
		return
	}
	logrus.Info("config.LoadGameConfig OK")

	common.InitCron()

}

func Finish() {
	common.FinishCron()
}

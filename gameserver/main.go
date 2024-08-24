package main

import (
	"dsservices/gameserver/gameplayer"
	"dsservices/gameserver/realm"
	"dsservices/gameserver/team"
	"dsservices/kissnet"
	"dsservices/launch"
	"dsservices/store"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
)

func main() {
	launch.Init()
	logrus.Info("game server")

	err := store.InitStore()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Info("store.InitStore error")
		return
	}
	team.Init()

	grpcURL := os.Getenv("DS_DSC_GRPC_URI")
	if len(grpcURL) == 0 {
		logrus.WithFields(logrus.Fields{}).Info("len(grpcURL) == 0")
		return
	}
	err = realm.RealmGRPCInit(grpcURL)
	if err != nil {
		logrus.Info("realm.RealmGRPCInit error")
		return
	}
	port, err := strconv.Atoi(os.Getenv("DS_GS_PORT"))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Info("DS_GS_PORT error")
		return
	}
	event := kissnet.NewNetEvent()
	logrus.Info("acceptor start")
	gAcceptor, err := kissnet.AcceptorFactory(
		"tcp",
		port,
		&gameplayer.GGamePlayerCB,
	)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("AcceptorFactory error")
		return
	}

	gAcceptor.Run()
	event.EventLoop()
	gAcceptor.Close()
	launch.Finish()
}

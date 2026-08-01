package main

import (
	"dsservices/common"
	"dsservices/config"
	"dsservices/gameserver/gameplayer"
	"dsservices/gameserver/realm"
	"dsservices/gameserver/team"
	"dsservices/kissnet"
	"dsservices/store"
	"fmt"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
)

func main() {
	common.Init()
	logrus.Info("game server")

	err := store.InitStore()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Info("store.InitStore error")
		return
	}
	if err := team.Init(); err != nil {
		logrus.WithError(err).Error("team.Init error")
		return
	}

	grpcURL := os.Getenv("DS_DSC_GRPC_URI")
	if len(grpcURL) == 0 {
		logrus.Info("DS_DSC_GRPC_URI empty")
		return
	}
	err = realm.DSCGameGRPCInit(grpcURL)
	if err != nil {
		logrus.WithError(err).Error("realm.DSCGameGRPCInit error")
		return
	}
	realm.StartRealmSubscriber()

	// Sweep disconnected players past the realm grace window.
	graceSec := config.GameConfig.Realm.DisconnectGraceSec
	if graceSec <= 0 {
		graceSec = 60
	}
	sweepSpec := fmt.Sprintf("@every %ds", graceSec/2+1)
	common.GCron.AddFunc(sweepSpec, func() {
		gameplayer.GPlayerMgr.SweepDisconnected(common.GetNowTime().Unix(), graceSec)
	})
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
	common.Finish()
}

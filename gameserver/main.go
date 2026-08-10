package main

import (
	"dsservices/common"
	"dsservices/config"
	"dsservices/gameserver/gameplayer"
	"dsservices/gameserver/realm"
	"dsservices/gameserver/team"
	"github.com/zs2619/kissnet-go"
	"dsservices/store"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

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
	defer store.FinishStore()

	// 先建 EventLoop 并注入，保证后续 cron/pubsub 对 player 的改动都串行进主循环。
	event := kissnet.NewNetEvent()
	gameplayer.SetNetEvent(event)

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
		now := common.GetNowTime().Unix()
		event.ScheduleMsg(func() {
			gameplayer.GPlayerMgr.SweepDisconnected(now, graceSec)
		})
	})
	port, err := strconv.Atoi(os.Getenv("DS_GS_PORT"))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Info("DS_GS_PORT error")
		return
	}
	logrus.Info("acceptor start")
	gAcceptor, err := kissnet.NewAcceptor(
		"tcp",
		port,
		&gameplayer.GGamePlayerCB,
		event,
	)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("NewAcceptor error")
		return
	}

	// acceptor 非阻塞启动，主线程留给 EventLoop（cellnet 风格）
	if err := gAcceptor.Start(); err != nil {
		logrus.WithError(err).Fatal("Acceptor Start error")
		return
	}

	// 信号触发优雅退出：先停 accept + 断连，等 OnDisconnectCB 入队，再关 EventLoop
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		logrus.Info("shutdown signal received")
		_ = gAcceptor.Close()
		if !gAcceptor.WaitConnectionsClosed(10 * time.Second) {
			logrus.Warn("not all connections closed within grace period")
		}
		event.Close()
	}()

	event.EventLoop()
	common.Finish()
}

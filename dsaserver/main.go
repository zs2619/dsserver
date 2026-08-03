package main

import (
	"dsservices/config"
	"dsservices/dsaserver/dsa"
	"dsservices/kissnet"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var g errgroup.Group

func main() {
	logrus.Info("dsa server start")
	_, err := config.LoadGameConfig("assets/config.json")
	if err != nil {
		logrus.WithError(err).Error("config.LoadGameConfig error")
		return
	}
	logrus.Info("config.LoadGameConfig OK")

	grpcURL := os.Getenv("DS_DSC_GRPC_URI")
	if len(grpcURL) == 0 {
		logrus.Error("DS_DSC_GRPC_URI empty")
		return
	}

	port, err := strconv.Atoi(os.Getenv("DS_DSA_PORT"))
	if err != nil {
		logrus.WithError(err).Error("DS_DSA_PORT error")
		return
	}
	agentID := os.Getenv("DS_DSA_AgentID")
	if len(agentID) == 0 {
		logrus.Error("DS_DSA_AgentID empty")
		return
	}

	grpcPort, err := strconv.Atoi(os.Getenv("DS_DSA_GRPC_PORT"))
	if err != nil {
		logrus.WithError(err).Error("DS_DSA_GRPC_PORT error")
		return
	}
	advertise := os.Getenv("DS_DSA_GRPC_ADDR")
	if advertise == "" {
		advertise = "127.0.0.1:" + strconv.Itoa(grpcPort)
	}

	reportClient, err := dsa.NewReportClient(agentID, grpcURL, advertise)
	if err != nil {
		logrus.WithError(err).Error("NewReportClient error")
		return
	}
	if err := reportClient.Register(); err != nil {
		logrus.WithError(err).Error("RegisterAgent error")
		return
	}

	agentServer := dsa.NewAgentServer(grpcPort)

	event := kissnet.NewNetEvent()
	logrus.Info("ds acceptor start")
	gAcceptor, err := kissnet.NewAcceptor(
		"tcp",
		port,
		&dsa.GDSAServerCB,
		event,
	)
	if err != nil {
		logrus.WithError(err).Fatal("NewAcceptor error")
		return
	}

	var shutdownOnce sync.Once
	shutdown := func() {
		shutdownOnce.Do(func() {
			// 停 CreateDS → 杀子进程并报 DEAD → 再关连接/EventLoop → 最后关 report。
			agentServer.Stop()
			dsa.GDSInfoMgr.ShutdownAll()
			_ = gAcceptor.Close()
			if !gAcceptor.WaitConnectionsClosed(10 * time.Second) {
				logrus.Warn("not all connections closed within grace period")
			}
			event.Close()
			reportClient.Close()
		})
	}

	g.Go(func() error {
		err := gAcceptor.Run()
		if err != nil {
			// accept 真错时主动拆掉其它组件，否则 EventLoop/gRPC 会卡住 g.Wait
			logrus.WithError(err).Error("acceptor run failed, shutting down")
			shutdown()
		}
		return err
	})
	g.Go(func() error {
		event.EventLoop()
		return nil
	})
	g.Go(func() error {
		return agentServer.Run()
	})
	g.Go(func() error {
		reportClient.RunHeartbeat()
		return nil
	})

	// 信号触发优雅退出
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		logrus.Info("shutdown signal received")
		shutdown()
	}()

	err = g.Wait()
	// 兜底清理（幂等）：信号/accept 错误路径可能已执行过
	shutdown()
	if err != nil {
		logrus.WithError(errors.WithStack(err)).Fatal("g.wait")
	}
	logrus.Info("dsa server end")
}

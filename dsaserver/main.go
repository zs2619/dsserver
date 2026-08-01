package main

import (
	"dsservices/config"
	"dsservices/dsaserver/dsa"
	"dsservices/kissnet"
	"os"
	"strconv"

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

	logrus.Info("ds acceptor start")
	gAcceptor, err := kissnet.AcceptorFactory(
		"tcp",
		port,
		&dsa.GDSAServerCB,
	)
	if err != nil {
		logrus.WithError(err).Fatal("AcceptorFactory error")
		return
	}

	g.Go(func() error {
		return gAcceptor.Run()
	})
	g.Go(func() error {
		return agentServer.Run()
	})
	g.Go(func() error {
		reportClient.RunHeartbeat()
		return nil
	})

	if err := g.Wait(); err != nil {
		logrus.WithError(errors.WithStack(err)).Fatal("g.wait")
	}
	reportClient.Close()
	logrus.Info("dsa server end")
}

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
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("config.LoadGameConfig error")
		return
	}
	logrus.Info("config.LoadGameConfig OK")

	grpcURL := os.Getenv("DS_DSC_GRPC_URI")
	if len(grpcURL) == 0 {
		logrus.WithFields(logrus.Fields{}).Error("len(grpcURL) == 0")
		return
	}

	port, err := strconv.Atoi(os.Getenv("DS_DSA_PORT"))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("DS_GS_PORT error")
		return
	}
	agentID := os.Getenv("DS_DSA_AgentID")
	if len(agentID) == 0 {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("DS_GS_PORT error")
		return
	}
	agent, err := dsa.NewDSAClient(agentID, grpcURL)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("dsa.NewDSAClient error")
	}

	logrus.Info("ds acceptor start")
	gAcceptor, err := kissnet.AcceptorFactory(
		"tcp",
		port,
		&dsa.GDSAServerCB,
	)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("AcceptorFactory error")
		return
	}
	g.Go(func() error {
		return gAcceptor.Run()
	})

	g.Go(func() error {
		return agent.RunStreamService()
	})
	if err := g.Wait(); err != nil {
		logrus.WithError(errors.WithStack(err)).Fatal("g.wait")
	}

	logrus.Info("dsa server end")
}

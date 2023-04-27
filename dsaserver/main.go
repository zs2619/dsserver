package main

import (
	"dsservices/config"
	"dsservices/dsaserver/dsa"
	"dsservices/kissnet"
	"dsservices/pb"

	"os"
	"strconv"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func main() {
	logrus.Info("dsa server")
	_, err := config.LoadGameConfig("assets/config.json")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Info("config.LoadGameConfig error")
		return
	}
	logrus.Info("config.LoadGameConfig OK")
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())
	// opts = append(opts, grpc.WithKeepaliveParams(keepalive.ClientParameters{
	// 	Time:                10 * time.Second,
	// 	Timeout:             100 * time.Millisecond,
	// 	PermitWithoutStream: true}))

	grpcURL := os.Getenv("DS_DSC_GRPC_URI")
	if len(grpcURL) == 0 {
		logrus.WithFields(logrus.Fields{}).Info("len(grpcURL) == 0")
		return
	}

	conn, err := grpc.Dial(grpcURL, opts...)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Info("dsa server error")
	}
	logrus.WithFields(
		logrus.Fields{
			"address": grpcURL,
		}).Info("connect dsc ok")

	client := pb.NewDsaDscARealmClient(conn)
	go dsa.RunWaitCreateRealm(client)

	port, err := strconv.Atoi(os.Getenv("DS_DSA_PORT"))
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
		dsa.DSAServerCB,
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
}

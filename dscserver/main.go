package main

import (
	"dsservices/common"
	"dsservices/config"
	"dsservices/dscserver/dsareport"
	"dsservices/dscserver/dsc"
	"dsservices/store"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var g errgroup.Group

func runHttpGM() error {
	http.HandleFunc("/createRealm", func(w http.ResponseWriter, req *http.Request) {})
	http.HandleFunc("/getInfo", func(w http.ResponseWriter, req *http.Request) {})
	http.HandleFunc("/stopAll", func(w http.ResponseWriter, req *http.Request) {})
	return http.ListenAndServe(":8888", nil)
}

func main() {
	logrus.Info("dsc server start")
	_, err := config.LoadGameConfig("assets/config.json")
	if err != nil {
		logrus.WithError(err).Fatal("config.LoadGameConfig error")
		return
	}
	if err := store.RedisInit(); err != nil {
		logrus.WithError(err).Fatal("RedisInit error")
		return
	}
	defer store.RedisFinish()

	common.InitCron()
	interval := config.GameConfig.Dsc.OrphanScanIntervalSec
	if interval <= 0 {
		interval = 30
	}
	_, _ = common.GCron.AddFunc(fmt.Sprintf("@every %ds", interval), dsareport.SweepOrphans)

	port, err := strconv.Atoi(os.Getenv("DS_DSC_PORT"))
	if err != nil {
		logrus.WithError(err).Fatal("DS_DSC_PORT error")
		return
	}

	dscServer, err := dsc.NewDSCServer(port)
	if err != nil {
		logrus.WithError(err).Fatal("dsc.NewDSCServer error")
		return
	}

	g.Go(func() error {
		dscServer.Run()
		return nil
	})
	g.Go(func() error {
		return runHttpGM()
	})

	if err := g.Wait(); err != nil {
		logrus.WithError(errors.WithStack(err)).Fatal("g.wait")
	}
	common.FinishCron()
	logrus.Info("dsc server end")
}

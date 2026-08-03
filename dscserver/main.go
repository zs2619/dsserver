package main

import (
	"context"
	"dsservices/common"
	"dsservices/config"
	"dsservices/dscserver/dsareport"
	"dsservices/dscserver/dsc"
	"dsservices/store"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

var g errgroup.Group

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
	defer common.FinishCron()
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

	mux := http.NewServeMux()
	mux.HandleFunc("/createRealm", func(w http.ResponseWriter, req *http.Request) {})
	mux.HandleFunc("/getInfo", func(w http.ResponseWriter, req *http.Request) {})
	mux.HandleFunc("/stopAll", func(w http.ResponseWriter, req *http.Request) {})
	httpServer := &http.Server{Addr: ":8888", Handler: mux}

	g.Go(func() error {
		err := dscServer.Run()
		if err == nil || err == grpc.ErrServerStopped {
			return nil
		}
		return err
	})
	g.Go(func() error {
		err := httpServer.ListenAndServe()
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	})

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		logrus.Info("shutdown signal received")
		dscServer.Stop()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(ctx)
	}()

	if err := g.Wait(); err != nil {
		logrus.WithError(errors.WithStack(err)).Fatal("g.wait")
	}
	logrus.Info("dsc server end")
}

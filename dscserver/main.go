package main

import (
	"net/http"
	"os"
	"strconv"

	"dsservices/dscserver/dsc"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var g errgroup.Group

func runHttpGM() error {
	createRealmHandler := func(w http.ResponseWriter, req *http.Request) {
	}
	getInfoHandler := func(w http.ResponseWriter, req *http.Request) {
	}

	stopAllHandler := func(w http.ResponseWriter, req *http.Request) {
	}

	http.HandleFunc("/createRealm", createRealmHandler)
	http.HandleFunc("/getInfo", getInfoHandler)
	http.HandleFunc("/stopAll", stopAllHandler)
	return http.ListenAndServe(":8888", nil)
}

func main() {
	logrus.Info("dsc server start")

	port, err := strconv.Atoi(os.Getenv("DS_DSC_PORT"))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Info("DS_GS_PORT error")
		return
	}

	dscServer, err := dsc.NewDSCServer(port)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Info("dsc.NewDSCServer error")
		return
	}

	g.Go(func() error {
		dscServer.Run()
		return nil
	})
	g.Go(func() error {
		runHttpGM()
		return nil
	})

	if err := g.Wait(); err != nil {
		logrus.WithError(errors.WithStack(err)).Fatal("g.wait")
	}
	logrus.Info("dsc server end")
}

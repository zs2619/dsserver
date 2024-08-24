package main

import (
	"flag"
	"net/http"
	"os"
	"strconv"

	"dsservices/dscserver/dsc"

	"github.com/sirupsen/logrus"
)

func initHttpGM() error {
	go func() {
		createRealmHandler := func(w http.ResponseWriter, req *http.Request) {
		}
		getInfoHandler := func(w http.ResponseWriter, req *http.Request) {
		}

		stopAllHandler := func(w http.ResponseWriter, req *http.Request) {
		}

		http.HandleFunc("/createRealm", createRealmHandler)
		http.HandleFunc("/getInfo", getInfoHandler)
		http.HandleFunc("/stopAll", stopAllHandler)
		http.ListenAndServe(":8888", nil)
	}()
	return nil
}

func main() {
	logrus.Info("dsc server")
	flag.Parse()

	port, err := strconv.Atoi(os.Getenv("DS_DSC_PORT"))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Info("DS_GS_PORT error")
		return
	}

	initHttpGM()
	dscServer, err := dsc.NewDSCServer(port)
	dscServer.Run()

}

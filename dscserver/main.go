package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	"google.golang.org/grpc"

	"dsservices/dscserver/dasdsc"
	"dsservices/dscserver/gamedsc"
	"dsservices/pb"

	"github.com/sirupsen/logrus"
)

func initHttpGM() error {
	go func() {
		createRealmHandler := func(w http.ResponseWriter, req *http.Request) {
			dasdsc.CreateRealmChan <- &pb.RpcCreateRealmInfo{RealmCfgID: "1"}
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

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterGameDscRealmServer(grpcServer, &gamedsc.RPCGameDscServer{})
	pb.RegisterDsaDscARealmServer(grpcServer, &dasdsc.RPCDasDscServer{})
	grpcServer.Serve(lis)
}

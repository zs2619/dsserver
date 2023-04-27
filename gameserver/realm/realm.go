package realm

import (
	"dsservices/pb"
	"context"

	"github.com/sirupsen/logrus"
)

func CreateRealm(realmCfgID string) error {
	_, err := GetRpcClient().CreateRealm(context.TODO(), &pb.RpcCreateRealmReq{})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Info("CreateRealm")
	}
	return nil
}

func QuickJoinRealm(teamID, realmCfgID string) (*pb.RpcQuickJoinRealmResp, error) {
	///TODO考虑并发调用问题
	resp, err := GetRpcClient().QuickJoinRealm(context.TODO(), &pb.RpcQuickJoinRealmReq{RealmCfgID: realmCfgID, TeamID: teamID})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Info("QuickJoinRealm")
		return nil, err
	}
	if !resp.Pending {
		//通知客户端加入副本
		notifyPlayerJoinRealm(resp.TeamID, resp.DsAddr, resp.RealmCfgID)
	}

	return resp, nil
}

func QueryRealmList() error {
	_, err := GetRpcClient().QueryRealmList(context.TODO(), &pb.RpcQueryRealmListRealmReq{})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Info("QueryRealmList")
	}
	return nil
}

type PlayerJoinRealm struct {
	TeamID, DSAddr, RealmCfgID string
}

var PlayerJoinRealmChan chan *PlayerJoinRealm = make(chan *PlayerJoinRealm, 2048)

func notifyPlayerJoinRealm(teamID, dsAddr, realmCfgID string) {
	PlayerJoinRealmChan <- &PlayerJoinRealm{TeamID: teamID, DSAddr: dsAddr, RealmCfgID: realmCfgID}
}

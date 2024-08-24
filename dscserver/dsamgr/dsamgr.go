package dsamgr

import (
	"dsservices/pb"
	"sync"
)

type DSAInfo struct {
	AgentName             string
	Addr                  string
	StreamServerEventChan chan *pb.StreamServerEvent
}

type DSAMgr struct {
	DSInfoMap sync.Map
}

func NewDSAInfo(AgentName, Addr string) *DSAInfo {
	info := &DSAInfo{StreamServerEventChan: make(chan *pb.StreamServerEvent)}
	return info
}

var GDSAMgr DSAMgr

func (mgr *DSAMgr) Add(dsa *DSAInfo) {
	mgr.DSInfoMap.Store(dsa.AgentName, dsa)
}
func (mgr *DSAMgr) Get(agentName string) *DSAInfo {
	dsa, ok := mgr.DSInfoMap.Load(agentName)
	if !ok {
		return nil
	}
	return dsa.(*DSAInfo)
}

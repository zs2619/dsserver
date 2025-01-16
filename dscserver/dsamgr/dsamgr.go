package dsamgr

import (
	"dsservices/pb"
	"sync"
)

type DSAMgr struct {
	DSInfoMap sync.Map
}

func NewDSAInfo(AgentName, Addr string, stream pb.StreamDscDSA_StreamServiceServer) *DSAInfo {
	info := &DSAInfo{
		StreamServerEventChan: make(chan *pb.StreamServerEvent),
		stream:                stream,
	}
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

func (mgr *DSAMgr) GetOptiamlDsa() *DSAInfo {
	return nil
}

func (mgr *DSAMgr) GetDsaByID(dsID string) *DSAInfo {
	return nil
}

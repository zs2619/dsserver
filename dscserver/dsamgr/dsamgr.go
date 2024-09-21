package dsamgr

import (
	"dsservices/pb"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type DSProc struct {
	Pid        int
	Cmd        string
	CfgRealmID string
}
type DSAInfo struct {
	AgentName             string
	Addr                  string
	StreamServerEventChan chan *pb.StreamServerEvent
	DSProcList            []*DSProc
	stream                pb.DsaDscADS_StreamServiceServer
}

func generateDSID() string {
	return ""
}

func (dsa *DSAInfo) AddDS(realmCgcID string) {
	dsID := generateDSID()
	creatDs := pb.RpcCreateDSReq{DsID: dsID, RealmCfgID: realmCgcID}
	dsa.Send2DSA(&creatDs)
}
func (dsa *DSAInfo) DelDS(dsID string) {
	removeDS := pb.RpcRemoveDSReq{DsID: dsID}
	dsa.Send2DSA(&removeDS)
}

func (dsa *DSAInfo) Send2DSA(resp proto.Message) (err error) {
	serverEvent := &pb.StreamServerEvent{}
	serverEvent.SEvent, err = anypb.New(resp)
	if err != nil {
		return
	}
	logrus.WithFields(logrus.Fields{"sevent": serverEvent.SEvent, "resp": resp}).Info("Send2DSA")
	dsa.StreamServerEventChan <- serverEvent
	return
}

func (dsa *DSAInfo) Count() int {
	return len(dsa.DSProcList)
}

func (dsa *DSAInfo) Run() {
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		for {
			data, err := dsa.stream.Recv()
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
					"data":  data,
				}).Info("WaitCreateDS error")
				break
			}
		}
		wg.Done()
	}()

	go func() {
		for info := range dsa.StreamServerEventChan {
			if info == nil {
				break
			}
			dsa.stream.Send(info)
			time.Sleep(time.Microsecond * 100)
		}
		wg.Done()
	}()

	wg.Wait()
}

func (dsa *DSAInfo) Close() {

}

type DSAMgr struct {
	DSInfoMap sync.Map
}

func NewDSAInfo(AgentName, Addr string, stream pb.DsaDscADS_StreamServiceServer) *DSAInfo {
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

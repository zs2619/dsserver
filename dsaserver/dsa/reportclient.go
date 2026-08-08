package dsa

import (
	"context"
	"dsservices/config"
	"dsservices/pb"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var GReportClient *ReportClient

type ReportClient struct {
	agentID      string
	grpcAddr     string // this DSA advertise addr for DSC callbacks
	dscURI       string
	conn         *grpc.ClientConn
	client       pb.RpcDsaReportClient
	quit         chan struct{}
	closeOnce    sync.Once
	capacity     int32
	runtime      string
	templateReady int32 // atomic bool
}

func NewReportClient(agentID, dscURI, advertiseAddr string) (*ReportClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, dscURI,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, err
	}
	rc := &ReportClient{
		agentID:  agentID,
		grpcAddr: advertiseAddr,
		dscURI:   dscURI,
		conn:     conn,
		client:   pb.NewRpcDsaReportClient(conn),
		capacity: config.GameConfig.Dsa.Capacity,
		runtime:  config.GameConfig.DSRuntime,
		quit:     make(chan struct{}),
	}
	GReportClient = rc
	return rc, nil
}

// SetTemplateReady 更新模板就绪状态（模板连接建立/断开时调用）。
func (rc *ReportClient) SetTemplateReady(ready bool) {
	if ready {
		atomic.StoreInt32(&rc.templateReady, 1)
	} else {
		atomic.StoreInt32(&rc.templateReady, 0)
	}
}

func (rc *ReportClient) isTemplateReady() bool {
	return atomic.LoadInt32(&rc.templateReady) != 0
}

func (rc *ReportClient) Register() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := rc.client.RegisterAgent(ctx, &pb.RegisterAgentReq{
		AgentID:       rc.agentID,
		GrpcAddr:      rc.grpcAddr,
		Capacity:      rc.capacity,
		Runtime:       rc.runtime,
		TemplateReady: rc.isTemplateReady(),
	})
	if err != nil {
		return err
	}
	if resp.Header != nil && resp.Header.Ret != pb.RpcHeader_OK {
		logrus.WithField("errmsg", resp.Header.Errmsg).Error("RegisterAgent rejected")
	}
	logrus.WithFields(logrus.Fields{"agentID": rc.agentID, "grpcAddr": rc.grpcAddr}).Info("RegisterAgent ok")
	return nil
}

func (rc *ReportClient) RunHeartbeat() {
	interval := config.GameConfig.Dsa.HeartbeatIntervalSec
	if interval <= 0 {
		interval = 3
	}
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-rc.quit:
			return
		case <-ticker.C:
		}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		_, err := rc.client.Heartbeat(ctx, &pb.AgentHeartbeatReq{
			AgentID:       rc.agentID,
			DsCount:       int32(GDSInfoMgr.Count()),
			TemplateReady: rc.isTemplateReady(),
		})
		cancel()
		if err != nil {
			logrus.WithError(err).Warn("Heartbeat failed")
		}
	}
}

func (rc *ReportClient) ReportDSEvent(req *pb.ReportDSEventReq) {
	if rc == nil {
		return
	}
	req.AgentID = rc.agentID
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := rc.client.ReportDSEvent(ctx, req)
	if err != nil {
		logrus.WithError(err).WithField("dsID", req.DsID).Error("ReportDSEvent")
	}
}

func (rc *ReportClient) Close() {
	rc.closeOnce.Do(func() {
		close(rc.quit)
		if rc.conn != nil {
			_ = rc.conn.Close()
		}
	})
}

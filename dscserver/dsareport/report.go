package dsareport

import (
	"context"
	"dsservices/config"
	"dsservices/dscserver/dsaclient"
	"dsservices/pb"
	"dsservices/store/dsastore"
	"time"

	"github.com/sirupsen/logrus"
)

type Server struct {
	pb.UnimplementedRpcDsaReportServer
}

func (s *Server) RegisterAgent(ctx context.Context, req *pb.RegisterAgentReq) (*pb.RegisterAgentResp, error) {
	if req.AgentID == "" || req.GrpcAddr == "" {
		return &pb.RegisterAgentResp{Header: &pb.RpcHeader{Ret: pb.RpcHeader_ReqParamError, Errmsg: "agentID/grpcAddr required"}}, nil
	}
	capacity := int64(req.Capacity)
	if capacity <= 0 {
		capacity = 100
	}
	runtime := req.Runtime
	if runtime == "" {
		runtime = "process"
	}
	if err := dsastore.RegisterAgent(ctx, req.AgentID, req.GrpcAddr, capacity, runtime, req.TemplateReady); err != nil {
		logrus.WithError(err).Error("RegisterAgent")
		return &pb.RegisterAgentResp{Header: &pb.RpcHeader{Ret: pb.RpcHeader_ReqParamError, Errmsg: err.Error()}}, nil
	}
	logrus.WithFields(logrus.Fields{"agentID": req.AgentID, "grpcAddr": req.GrpcAddr, "runtime": runtime, "templateReady": req.TemplateReady}).Info("RegisterAgent ok")
	return &pb.RegisterAgentResp{Header: &pb.RpcHeader{Ret: pb.RpcHeader_OK}}, nil
}

func (s *Server) Heartbeat(ctx context.Context, req *pb.AgentHeartbeatReq) (*pb.AgentHeartbeatResp, error) {
	if req.AgentID == "" {
		return &pb.AgentHeartbeatResp{Header: &pb.RpcHeader{Ret: pb.RpcHeader_ReqParamError}}, nil
	}
	if err := dsastore.HeartbeatAgent(ctx, req.AgentID, int64(req.DsCount), req.TemplateReady); err != nil {
		return &pb.AgentHeartbeatResp{Header: &pb.RpcHeader{Ret: pb.RpcHeader_ReqParamError, Errmsg: err.Error()}}, nil
	}
	return &pb.AgentHeartbeatResp{Header: &pb.RpcHeader{Ret: pb.RpcHeader_OK}}, nil
}

func (s *Server) ReportDSEvent(ctx context.Context, req *pb.ReportDSEventReq) (*pb.ReportDSEventResp, error) {
	if req.DsID == "" {
		return &pb.ReportDSEventResp{Header: &pb.RpcHeader{Ret: pb.RpcHeader_ReqParamError}}, nil
	}
	ds, err := dsastore.GetDS(ctx, req.DsID)
	if err != nil {
		return &pb.ReportDSEventResp{Header: &pb.RpcHeader{Ret: pb.RpcHeader_ReqParamError, Errmsg: err.Error()}}, nil
	}
	if ds == nil {
		return &pb.ReportDSEventResp{Header: &pb.RpcHeader{Ret: pb.RpcHeader_ReqParamError, Errmsg: "ds not found"}}, nil
	}
	realmID := req.RealmID
	if realmID == "" {
		realmID = ds.RealmID
	}

	switch req.EventType {
	case pb.DSEventType_DS_EVENT_READY:
		// DS 已非 creating：忽略（超时/死）。
		if ds.State != dsastore.DSStateCreating && ds.State != dsastore.DSStateReady {
			logrus.WithFields(logrus.Fields{"dsID": req.DsID, "state": ds.State}).Warn("ignore READY in non-creating ds state")
			return &pb.ReportDSEventResp{Header: &pb.RpcHeader{Ret: pb.RpcHeader_OK}}, nil
		}
		// realm ended/fail：拒绝复活；已 ready：幂等返回。
		allow, alreadyReady, err := dsastore.RealmAllowsReady(ctx, realmID)
		if err != nil {
			return &pb.ReportDSEventResp{Header: &pb.RpcHeader{Ret: pb.RpcHeader_ReqParamError, Errmsg: err.Error()}}, nil
		}
		if !allow {
			logrus.WithFields(logrus.Fields{"dsID": req.DsID, "realmID": realmID}).Warn("ignore READY: realm ended/fail")
			go removeRemoteDS(ds)
			return &pb.ReportDSEventResp{Header: &pb.RpcHeader{Ret: pb.RpcHeader_OK}}, nil
		}
		addr := req.DsAddr
		if addr == "" {
			addr = ds.DSAddr
		}
		if addr == "" {
			return &pb.ReportDSEventResp{Header: &pb.RpcHeader{Ret: pb.RpcHeader_ReqParamError, Errmsg: "dsAddr required"}}, nil
		}
		if alreadyReady {
			return &pb.ReportDSEventResp{Header: &pb.RpcHeader{Ret: pb.RpcHeader_OK}}, nil
		}
		if err := dsastore.UpdateDSState(ctx, req.DsID, dsastore.DSStateReady, addr); err != nil {
			return &pb.ReportDSEventResp{Header: &pb.RpcHeader{Ret: pb.RpcHeader_ReqParamError, Errmsg: err.Error()}}, nil
		}
		if err := dsastore.MarkRealmReady(ctx, realmID, req.DsID, addr); err != nil {
			return &pb.ReportDSEventResp{Header: &pb.RpcHeader{Ret: pb.RpcHeader_ReqParamError, Errmsg: err.Error()}}, nil
		}
		logrus.WithFields(logrus.Fields{"dsID": req.DsID, "realmID": realmID, "dsAddr": addr}).Info("DS CreateOK → realm ready")

	case pb.DSEventType_DS_EVENT_PLAYER_ENTER, pb.DSEventType_DS_EVENT_START:
		_ = dsastore.UpdateDSState(ctx, req.DsID, dsastore.DSStatePlaying, "")
		_ = dsastore.MarkRealmInGame(ctx, realmID)

	case pb.DSEventType_DS_EVENT_GAME_END:
		_ = dsastore.UpdateDSState(ctx, req.DsID, dsastore.DSStateEnded, "")
		_ = dsastore.MarkRealmEnded(ctx, realmID)
		_ = dsastore.ClearRealmUserMappings(ctx, realmID)
		go removeRemoteDS(ds)

	case pb.DSEventType_DS_EVENT_DEAD:
		_ = dsastore.UpdateDSState(ctx, req.DsID, dsastore.DSStateDead, "")
		_ = dsastore.MarkRealmFail(ctx, realmID, req.FailReason)
		_ = dsastore.ClearRealmUserMappings(ctx, realmID)
		_ = dsastore.ReleaseRealmAlloc(ctx, realmID)
		go removeRemoteDS(ds)

	default:
		return &pb.ReportDSEventResp{Header: &pb.RpcHeader{Ret: pb.RpcHeader_ReqParamError, Errmsg: "unknown event"}}, nil
	}
	return &pb.ReportDSEventResp{Header: &pb.RpcHeader{Ret: pb.RpcHeader_OK}}, nil
}

func removeRemoteDS(ds *dsastore.DSInfo) {
	if ds == nil {
		return
	}
	ctx := context.Background()
	agent, err := dsastore.GetAgent(ctx, ds.AgentID)
	if err != nil || agent == nil {
		return
	}
	_, _ = dsaclient.RemoveDS(ctx, agent.GrpcAddr, &pb.DsaRemoveDSReq{DsID: ds.DSID})
	_ = dsastore.DeleteDSMeta(ctx, ds.DSID)
}

// SweepOrphans marks stuck creating DS / stale agents.
func SweepOrphans() {
	ctx := context.Background()
	hbTimeout := config.GameConfig.Dsc.AgentHeartbeatTimeoutSec
	createTimeout := config.GameConfig.Dsc.DsCreatingTimeoutSec
	readyTimeout := config.GameConfig.Dsc.DsReadyIdleTimeoutSec
	playTimeout := config.GameConfig.Dsc.DsPlayingIdleTimeoutSec

	stale, err := dsastore.ListStaleAgents(ctx, hbTimeout)
	if err == nil {
		for _, agentID := range stale {
			logrus.WithField("agentID", agentID).Warn("agent heartbeat timeout")
			_ = dsastore.MarkAgentOffline(ctx, agentID)
			dslist, _ := dsastore.ListDSByAgent(ctx, agentID)
			for _, ds := range dslist {
				_ = dsastore.UpdateDSState(ctx, ds.DSID, dsastore.DSStateDead, "")
				_ = dsastore.MarkRealmFail(ctx, ds.RealmID, "agent offline")
				_ = dsastore.ClearRealmUserMappings(ctx, ds.RealmID)
				_ = dsastore.DeleteDSMeta(ctx, ds.DSID)
			}
		}
	}

	now := time.Now().Unix()
	list, err := dsastore.ListCreatingOrPlayingDS(ctx)
	if err != nil {
		return
	}
	for _, ds := range list {
		age := now - ds.UpdatedAt
		if ds.State == dsastore.DSStateCreating && age > createTimeout {
			logrus.WithFields(logrus.Fields{"dsID": ds.DSID, "realmID": ds.RealmID, "ageSec": age}).Warn("ds creating timeout (no CreateOK)")
			_ = dsastore.UpdateDSState(ctx, ds.DSID, dsastore.DSStateDead, "")
			_ = dsastore.MarkRealmFail(ctx, ds.RealmID, "creating timeout")
			_ = dsastore.ClearRealmUserMappings(ctx, ds.RealmID)
			_ = dsastore.ReleaseRealmAlloc(ctx, ds.RealmID)
			go removeRemoteDS(ds)
			continue
		}
		if ds.State == dsastore.DSStateReady && readyTimeout > 0 && age > readyTimeout {
			logrus.WithFields(logrus.Fields{"dsID": ds.DSID, "realmID": ds.RealmID, "ageSec": age}).Warn("ds ready idle timeout (no player enter)")
			_ = dsastore.UpdateDSState(ctx, ds.DSID, dsastore.DSStateDead, "")
			_ = dsastore.MarkRealmFail(ctx, ds.RealmID, "ready idle timeout")
			_ = dsastore.ClearRealmUserMappings(ctx, ds.RealmID)
			_ = dsastore.ReleaseRealmAlloc(ctx, ds.RealmID)
			go removeRemoteDS(ds)
			continue
		}
		if ds.State == dsastore.DSStatePlaying && age > playTimeout {
			logrus.WithField("dsID", ds.DSID).Warn("ds playing idle timeout")
			_ = dsastore.UpdateDSState(ctx, ds.DSID, dsastore.DSStateEnded, "")
			_ = dsastore.MarkRealmEnded(ctx, ds.RealmID)
			_ = dsastore.ClearRealmUserMappings(ctx, ds.RealmID)
			go removeRemoteDS(ds)
		}
	}

	// Network-partition / CleanupRealm races may leave stub realm hashes.
	staleRealm := createTimeout
	if readyTimeout > staleRealm {
		staleRealm = readyTimeout
	}
	if n, err := dsastore.SweepRealmFragments(ctx, createTimeout, staleRealm); err != nil {
		logrus.WithError(err).Warn("SweepRealmFragments")
	} else if n > 0 {
		logrus.WithField("cleaned", n).Info("realm fragments cleaned")
	}
}

package gamedsc

import (
	"context"
	"dsservices/config"
	"dsservices/dscserver/dsaclient"
	"dsservices/pb"
	"dsservices/store/dsastore"

	"github.com/sirupsen/logrus"
)

type RPCGameDscServer struct {
	pb.UnimplementedRpcGameDscServer
}

func (gdServer *RPCGameDscServer) CreateRealm(ctx context.Context, req *pb.RpcCreateRealmReq) (*pb.RpcCreateRealmResp, error) {
	if req.RealmID == "" || req.RealmCfgID == "" {
		return &pb.RpcCreateRealmResp{
			Header: &pb.RpcHeader{Ret: pb.RpcHeader_ReqParamError, Errmsg: "realmID/realmCfgID required"},
			Status: pb.CreateRealmStatus_CREATE_REALM_FAILED,
		}, nil
	}

	claim, dsAddr, err := dsastore.TryClaimRealmAlloc(ctx, req.RealmID)
	if err != nil {
		return &pb.RpcCreateRealmResp{
			Header: &pb.RpcHeader{Ret: pb.RpcHeader_ReqParamError, Errmsg: err.Error()},
			Status: pb.CreateRealmStatus_CREATE_REALM_FAILED,
		}, nil
	}
	switch claim {
	case dsastore.AllocClaimReady:
		return &pb.RpcCreateRealmResp{
			Header: &pb.RpcHeader{Ret: pb.RpcHeader_OK},
			Status: pb.CreateRealmStatus_CREATE_REALM_ALREADY_READY,
			DsAddr: dsAddr,
		}, nil
	case dsastore.AllocClaimCreating:
		return &pb.RpcCreateRealmResp{
			Header: &pb.RpcHeader{Ret: pb.RpcHeader_OK},
			Status: pb.CreateRealmStatus_CREATE_REALM_ALREADY_CREATING,
		}, nil
	case dsastore.AllocClaimMissing:
		// GS already CleanupRealm (e.g. local fail); do not recreate stub.
		return &pb.RpcCreateRealmResp{
			Header: &pb.RpcHeader{Ret: pb.RpcHeader_ReqParamError, Errmsg: "realm skeleton missing"},
			Status: pb.CreateRealmStatus_CREATE_REALM_FAILED,
		}, nil
	case dsastore.AllocClaimOK:
		// sole winner: async allocate
	default:
		return &pb.RpcCreateRealmResp{
			Header: &pb.RpcHeader{Ret: pb.RpcHeader_OK},
			Status: pb.CreateRealmStatus_CREATE_REALM_ALREADY_CREATING,
		}, nil
	}

	go allocateDS(req.RealmID, req.TeamID, req.RealmCfgID)

	return &pb.RpcCreateRealmResp{
		Header: &pb.RpcHeader{Ret: pb.RpcHeader_OK},
		Status: pb.CreateRealmStatus_CREATE_REALM_ACCEPTED,
	}, nil
}

func allocateDS(realmID, teamID, realmCfgID string) {
	ctx := context.Background()
	hbTimeout := config.GameConfig.Dsc.AgentHeartbeatTimeoutSec

	if !realmStillAllocating(ctx, realmID) {
		return
	}

	// already bound (retry / race after claim)
	if existing, err := dsastore.GetDSIDByRealm(ctx, realmID); err == nil && existing != "" {
		ds, _ := dsastore.GetDS(ctx, existing)
		if ds != nil && (ds.State == dsastore.DSStateReady || ds.State == dsastore.DSStatePlaying) {
			if realmStillAllocating(ctx, realmID) {
				_ = dsastore.MarkRealmReady(ctx, realmID, ds.DSID, ds.DSAddr)
			}
			return
		}
		if ds != nil && ds.State == dsastore.DSStateCreating {
			return
		}
	}

	if !realmStillAllocating(ctx, realmID) {
		return
	}

	agent, err := dsastore.ReserveOptimalAgent(ctx, hbTimeout)
	if err != nil || agent == nil {
		logrus.WithError(err).WithField("realmID", realmID).Error("no available agent")
		if realmStillAllocating(ctx, realmID) {
			_ = dsastore.MarkRealmFail(ctx, realmID, "no available agent")
			_ = dsastore.ReleaseRealmAlloc(ctx, realmID)
		}
		return
	}

	if !realmStillAllocating(ctx, realmID) {
		_ = dsastore.ReleaseAgentSlot(ctx, agent.AgentID)
		return
	}

	dsID := dsastore.GenerateDSID()
	if err := dsastore.CreateDSMeta(ctx, dsID, realmID, agent.AgentID, realmCfgID); err != nil {
		logrus.WithError(err).Error("CreateDSMeta")
		_ = dsastore.ReleaseAgentSlot(ctx, agent.AgentID)
		if realmStillAllocating(ctx, realmID) {
			_ = dsastore.MarkRealmFail(ctx, realmID, "create ds meta failed")
			_ = dsastore.ReleaseRealmAlloc(ctx, realmID)
		}
		return
	}

	if !realmStillAllocating(ctx, realmID) {
		abortInFlightDS(ctx, dsID, agent.GrpcAddr)
		return
	}

	resp, err := dsaclient.CreateDS(ctx, agent.GrpcAddr, &pb.DsaCreateDSReq{
		DsID:       dsID,
		RealmID:    realmID,
		RealmCfgID: realmCfgID,
	})
	if err != nil || resp == nil || resp.Header == nil || resp.Header.Ret != pb.RpcHeader_OK {
		errmsg := "CreateDS rpc failed"
		if err != nil {
			errmsg = err.Error()
		} else if resp != nil && resp.Header != nil && resp.Header.Errmsg != "" {
			errmsg = resp.Header.Errmsg
		}
		logrus.WithError(err).WithFields(logrus.Fields{"realmID": realmID, "dsID": dsID, "teamID": teamID}).Error(errmsg)
		_ = dsastore.UpdateDSState(ctx, dsID, dsastore.DSStateDead, "")
		if realmStillAllocating(ctx, realmID) {
			_ = dsastore.MarkRealmFail(ctx, realmID, errmsg)
			_ = dsastore.ReleaseRealmAlloc(ctx, realmID)
		}
		_ = dsastore.DeleteDSMeta(ctx, dsID) // also ReleaseAgentSlot
		return
	}

	if !realmStillAllocating(ctx, realmID) {
		abortInFlightDS(ctx, dsID, agent.GrpcAddr)
		return
	}

	// 进程已拉起，仍保持 creating；等 DSA ReportDSEvent(READY/CreateOK) 再 MarkRealmReady。
	if resp.DsAddr != "" {
		_ = dsastore.UpdateDSState(ctx, dsID, dsastore.DSStateCreating, resp.DsAddr)
	}
	logrus.WithFields(logrus.Fields{
		"realmID": realmID, "dsID": dsID, "agent": agent.AgentID, "dsAddr": resp.DsAddr,
	}).Info("CreateDS accepted, waiting DS CreateOK")
}

func realmStillAllocating(ctx context.Context, realmID string) bool {
	ok, err := dsastore.RealmAllowsAllocate(ctx, realmID)
	if err != nil {
		logrus.WithError(err).WithField("realmID", realmID).Warn("RealmAllowsAllocate")
		return false
	}
	if !ok {
		logrus.WithField("realmID", realmID).Info("allocate aborted: realm no longer creating")
	}
	return ok
}

// abortInFlightDS: RemoveRealm 等已结束分配时，回收已拉起的 DS。
func abortInFlightDS(ctx context.Context, dsID, grpcAddr string) {
	if grpcAddr != "" {
		_, _ = dsaclient.RemoveDS(ctx, grpcAddr, &pb.DsaRemoveDSReq{DsID: dsID})
	}
	_ = dsastore.UpdateDSState(ctx, dsID, dsastore.DSStateDead, "")
	_ = dsastore.DeleteDSMeta(ctx, dsID)
}

func (gdServer *RPCGameDscServer) RemoveRealm(ctx context.Context, req *pb.RpcRemoveRealmReq) (*pb.RpcRemoveRealmResp, error) {
	if req.RealmID == "" {
		return &pb.RpcRemoveRealmResp{Header: &pb.RpcHeader{Ret: pb.RpcHeader_ReqParamError}}, nil
	}
	// 先 ended，阻断并发 allocateDS / 迟到 CreateOK 复活。
	_ = dsastore.MarkRealmEnded(ctx, req.RealmID)
	_ = dsastore.ReleaseRealmAlloc(ctx, req.RealmID)

	dsID, err := dsastore.GetDSIDByRealm(ctx, req.RealmID)
	if err == nil && dsID != "" {
		ds, _ := dsastore.GetDS(ctx, dsID)
		if ds != nil {
			agent, _ := dsastore.GetAgent(ctx, ds.AgentID)
			if agent != nil {
				_, _ = dsaclient.RemoveDS(ctx, agent.GrpcAddr, &pb.DsaRemoveDSReq{DsID: dsID})
			}
			_ = dsastore.DeleteDSMeta(ctx, dsID)
		}
	}
	_ = dsastore.ClearRealmUserMappings(ctx, req.RealmID)
	return &pb.RpcRemoveRealmResp{Header: &pb.RpcHeader{Ret: pb.RpcHeader_OK}}, nil
}

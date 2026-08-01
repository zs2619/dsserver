package realm

import (
	"context"
	"dsservices/pb"
	"dsservices/store/dsastore"
	"time"

	"github.com/sirupsen/logrus"
)

// CreateRealmAndWatch calls DSC CreateRealm then watches Redis for ready/fail.
func CreateRealmAndWatch(realmID, teamID, realmCfgID string, userIDs []string) {
	NotifyWatchRealm(realmID, teamID, realmCfgID, userIDs)

	client := GetRpcClient()
	if client == nil {
		logrus.WithField("realmID", realmID).Error("CreateRealm: dsc client nil")
		failRealmLocal(realmID, userIDs, "dsc client nil")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := client.CreateRealm(ctx, &pb.RpcCreateRealmReq{
		RealmID:    realmID,
		TeamID:     teamID,
		RealmCfgID: realmCfgID,
	})
	if err != nil {
		logrus.WithError(err).WithField("realmID", realmID).Error("CreateRealm rpc")
		failRealmLocal(realmID, userIDs, "CreateRealm rpc failed")
		return
	}
	if resp == nil {
		failRealmLocal(realmID, userIDs, "CreateRealm empty resp")
		return
	}
	switch resp.Status {
	case pb.CreateRealmStatus_CREATE_REALM_ALREADY_READY:
		PushJoinRealmOK(userIDs, realmID, teamID, realmCfgID, resp.DsAddr)
	case pb.CreateRealmStatus_CREATE_REALM_ACCEPTED, pb.CreateRealmStatus_CREATE_REALM_ALREADY_CREATING:
		logrus.WithFields(logrus.Fields{"realmID": realmID, "status": resp.Status}).Info("CreateRealm accepted")
	case pb.CreateRealmStatus_CREATE_REALM_FAILED:
		logrus.WithField("realmID", realmID).Warn("CreateRealm failed")
		failRealmLocal(realmID, userIDs, "CreateRealm rejected")
	}
}

// failRealmLocal handles CreateRealm failure on GS side (RPC error / rejected).
// Do NOT call MarkRealmFail here: it publishes realm:fail and onRealmFail would
// push JoinRealmFail again (double notify). CleanupRealm is enough; DSC allocateDS
// / SweepOrphans stop when realm state is gone or not creating.
// Only notify if we still owned the watch entry; if onRealmFail already fired
// (DSC-side fail raced ahead of our RPC timeout), it already notified + cleaned.
func failRealmLocal(realmID string, userIDs []string, reason string) {
	ctx := context.Background()
	watchMu.Lock()
	_, stillWatching := watching[realmID]
	delete(watching, realmID)
	watchMu.Unlock()
	if !stillWatching {
		return
	}
	if PushRealmFailToLocalPlayers != nil {
		PushRealmFailToLocalPlayers(userIDs, realmID)
	}
	_ = dsastore.CleanupRealm(ctx, realmID)
	logrus.WithFields(logrus.Fields{"realmID": realmID, "reason": reason}).Warn("realm fail notified locally")
}

// RemoveRealm asks DSC to tear down a realm (and its ds).
func RemoveRealm(realmID string) {
	client := GetRpcClient()
	if client == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := client.RemoveRealm(ctx, &pb.RpcRemoveRealmReq{RealmID: realmID})
	if err != nil {
		logrus.WithError(err).WithField("realmID", realmID).Error("RemoveRealm rpc")
	}
}

// TryPushRealmOnLogin: reconnect / switch GS ? read realm mapping and push if ready.
// onReady sets local player realm/team before JoinRealmOK.
func TryPushRealmOnLogin(userID string, onReady func(realmID, teamID, realmCfgID, dsAddr string)) {
	ctx := context.Background()
	realmID, err := dsastore.GetRealmIDByUser(ctx, userID)
	if err != nil || realmID == "" {
		return
	}
	info, err := dsastore.GetRealm(ctx, realmID)
	if err != nil || info == nil {
		return
	}
	switch info.State {
	case dsastore.RealmStateReady, dsastore.RealmStateInGame:
		if info.DSAddr != "" && onReady != nil {
			onReady(realmID, info.TeamID, info.RealmCfgID, info.DSAddr)
		}
	case dsastore.RealmStateCreating:
		// Only watch this reconnecting user ? avoids bloating every GS watch map
		// with the full roster; pubsub payload still delivers to other local players.
		NotifyWatchRealm(realmID, info.TeamID, info.RealmCfgID, []string{userID})
	case dsastore.RealmStateFail:
		_ = dsastore.ClearRealmUserMappings(ctx, realmID)
	}
}

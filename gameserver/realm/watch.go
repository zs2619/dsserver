package realm

import (
	"context"
	"dsservices/store"
	"dsservices/store/dsastore"
	"encoding/json"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// PushToLocalPlayers is set by gameplayer to avoid import cycles.
var PushToLocalPlayers func(userIDs []string, realmID, teamID, realmCfgID, dsAddr string)

// PushRealmFailToLocalPlayers notifies local players their realm failed (set by gameplayer).
var PushRealmFailToLocalPlayers func(userIDs []string, realmID string)

// LocalPlayerExists is set by gameplayer; used to skip Redis work on unrelated GS.
var LocalPlayerExists func(userID string) bool

type watchEntry struct {
	teamID     string
	realmCfgID string
	userIDs    []string
}

var (
	watchMu  sync.Mutex
	watching = make(map[string]*watchEntry)
	subOnce  sync.Once
)

func PushJoinRealmOK(userIDs []string, realmID, teamID, realmCfgID, dsAddr string) {
	if PushToLocalPlayers != nil {
		PushToLocalPlayers(userIDs, realmID, teamID, realmCfgID, dsAddr)
	}
}

// NotifyWatchRealm registers interest in a realm's ready/fail pubsub.
// Merges userIDs so reconnect only adds the local user without wiping the create-side list.
func NotifyWatchRealm(realmID, teamID, realmCfgID string, userIDs []string) {
	watchMu.Lock()
	if prev, ok := watching[realmID]; ok && prev != nil {
		if teamID != "" {
			prev.teamID = teamID
		}
		if realmCfgID != "" {
			prev.realmCfgID = realmCfgID
		}
		prev.userIDs = mergeUserIDs(prev.userIDs, userIDs)
	} else {
		watching[realmID] = &watchEntry{teamID: teamID, realmCfgID: realmCfgID, userIDs: append([]string{}, userIDs...)}
	}
	watchMu.Unlock()
	ensureSubscriber()
}

func mergeUserIDs(a, b []string) []string {
	seen := make(map[string]struct{}, len(a)+len(b))
	out := make([]string, 0, len(a)+len(b))
	for _, uid := range a {
		if uid == "" {
			continue
		}
		if _, ok := seen[uid]; ok {
			continue
		}
		seen[uid] = struct{}{}
		out = append(out, uid)
	}
	for _, uid := range b {
		if uid == "" {
			continue
		}
		if _, ok := seen[uid]; ok {
			continue
		}
		seen[uid] = struct{}{}
		out = append(out, uid)
	}
	return out
}

func ensureSubscriber() {
	subOnce.Do(func() {
		go runSubscriber()
	})
}

func StartRealmSubscriber() {
	ensureSubscriber()
}

func runSubscriber() {
	for {
		subscribeLoop()
		// Redis connection dropped; back off before reconnecting.
		logrus.Warn("realm pubsub channel closed, reconnecting in 1s")
		time.Sleep(time.Second)
	}
}

func subscribeLoop() {
	pubsub := store.RedisClient.PSubscribe(context.Background(), "realm:ready:*", "realm:fail:*")
	defer pubsub.Close()
	ch := pubsub.Channel()
	logrus.Info("realm pubsub subscriber started")
	for msg := range ch {
		channel := msg.Channel
		payload := msg.Payload
		realmID := extractRealmID(channel)
		if realmID == "" {
			continue
		}
		if hasPrefix(channel, "realm:ready:") {
			onRealmReady(realmID, payload)
		} else if hasPrefix(channel, "realm:fail:") {
			onRealmFail(realmID, payload)
		}
	}
}

func extractRealmID(channel string) string {
	for _, prefix := range []string{"realm:ready:", "realm:fail:"} {
		if hasPrefix(channel, prefix) {
			return channel[len(prefix):]
		}
	}
	return ""
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func parsePubPayload(raw string) dsastore.RealmPubPayload {
	var p dsastore.RealmPubPayload
	if raw == "" {
		return p
	}
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		// backward compat: plain dsAddr / reason string
		p.DsAddr = raw
		p.Reason = raw
	}
	return p
}

func anyLocalPlayer(userIDs []string) bool {
	if LocalPlayerExists == nil {
		return true
	}
	for _, uid := range userIDs {
		if LocalPlayerExists(uid) {
			return true
		}
	}
	return false
}

func onRealmReady(realmID, raw string) {
	pub := parsePubPayload(raw)

	watchMu.Lock()
	entry, watched := watching[realmID]
	delete(watching, realmID)
	watchMu.Unlock()

	// Early-out when pubsub carries userIDs (or we have a watch entry) and none are local.
	candidates := pub.UserIDs
	if len(candidates) == 0 && watched && entry != nil {
		candidates = entry.userIDs
	}
	if len(candidates) > 0 && !anyLocalPlayer(candidates) {
		return
	}

	ctx := context.Background()
	info, err := dsastore.GetRealm(ctx, realmID)
	if err != nil || info == nil {
		return
	}
	addr := pub.DsAddr
	if addr == "" {
		addr = info.DSAddr
	}
	teamID := info.TeamID
	if teamID == "" {
		teamID = pub.TeamID
	}
	realmCfgID := info.RealmCfgID
	if realmCfgID == "" {
		realmCfgID = pub.RealmCfgID
	}
	userIDs := info.UserIDList
	if len(userIDs) == 0 {
		userIDs = pub.UserIDs
	}
	if len(userIDs) == 0 && watched && entry != nil {
		userIDs = entry.userIDs
		if teamID == "" {
			teamID = entry.teamID
		}
		if realmCfgID == "" {
			realmCfgID = entry.realmCfgID
		}
	}
	if !anyLocalPlayer(userIDs) {
		return
	}
	PushJoinRealmOK(userIDs, realmID, teamID, realmCfgID, addr)
}

func onRealmFail(realmID, raw string) {
	pub := parsePubPayload(raw)

	watchMu.Lock()
	entry, watched := watching[realmID]
	delete(watching, realmID)
	watchMu.Unlock()

	candidates := pub.UserIDs
	if len(candidates) == 0 && watched && entry != nil {
		candidates = entry.userIDs
	}
	hasLocal := anyLocalPlayer(candidates)
	ctx := context.Background()
	// Unrelated GS: skip. Convert-side (watched) still cleans even if players moved away.
	if len(candidates) > 0 && !hasLocal && !watched {
		return
	}

	info, _ := dsastore.GetRealm(ctx, realmID)
	var userIDs []string
	if info != nil {
		userIDs = info.UserIDList
	}
	if len(userIDs) == 0 {
		userIDs = pub.UserIDs
	}
	if len(userIDs) == 0 && watched && entry != nil {
		userIDs = entry.userIDs
	}
	if PushRealmFailToLocalPlayers != nil && len(userIDs) > 0 && anyLocalPlayer(userIDs) {
		PushRealmFailToLocalPlayers(userIDs, realmID)
	}
	if watched || anyLocalPlayer(userIDs) {
		_ = dsastore.CleanupRealm(ctx, realmID)
	}
	logrus.WithField("realmID", realmID).Warn("realm fail notified and cleaned")
}

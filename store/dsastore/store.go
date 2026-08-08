package dsastore

import (
	"context"
	"dsservices/store"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RealmPubPayload is published on realm:ready / realm:fail for GS early-out filtering.
type RealmPubPayload struct {
	DsAddr     string   `json:"dsAddr,omitempty"`
	TeamID     string   `json:"teamID,omitempty"`
	RealmCfgID string   `json:"realmCfgID,omitempty"`
	UserIDs    []string `json:"userIDs,omitempty"`
	Reason     string   `json:"reason,omitempty"`
}

var (
	// claim: KEYS realmKey, realm2dsKey, realmIndex; ARGV now, realmID
	// → {ok|ready|creating|missing, dsAddr?}
	// missing: skeleton already cleaned — never recreate a stub hash.
	tryClaimRealmAllocScript = redis.NewScript(`
local realmKey, realm2dsKey, realmIndex = KEYS[1], KEYS[2], KEYS[3]
local now, realmID = ARGV[1], ARGV[2]
if redis.call('exists', realmKey) == 0 then
  return {'missing', ''}
end
local fields = redis.call('hmget', realmKey, 'state', 'dsAddr', 'dsID', 'alloc')
local state, dsAddr, dsID, alloc = fields[1], fields[2], fields[3], fields[4]
if state == 'ready' or state == 'ingame' then
  return {'ready', dsAddr or ''}
end
local bound = redis.call('get', realm2dsKey)
if (bound and bound ~= false and bound ~= '') or (dsID and dsID ~= '') or alloc == '1' then
  return {'creating', ''}
end
if state == 'fail' or state == 'ended' then
  redis.call('hmset', realmKey, 'state', 'creating', 'alloc', '1', 'dsID', '', 'dsAddr', '', 'failReason', '', 'updatedAt', now)
  redis.call('del', realm2dsKey)
  redis.call('sadd', realmIndex, realmID)
  return {'ok', ''}
end
redis.call('hset', realmKey, 'alloc', '1', 'updatedAt', now)
if state == false or state == nil or state == '' then
  redis.call('hset', realmKey, 'state', 'creating')
end
redis.call('sadd', realmIndex, realmID)
return {'ok', ''}
`)

	// reserve: KEYS agentsZSet → {agentID, grpcAddr} or empty
	// ARGV now, hbTimeoutSec
	// 跳过 templateReady != "1" 的 Agent（模板未就绪不可 fork）。
	reserveOptimalAgentScript = redis.NewScript(`
local now, hbTimeout = tonumber(ARGV[1]), tonumber(ARGV[2])
local members = redis.call('zrange', KEYS[1], 0, 30)
for _, agentID in ipairs(members) do
  local akey = '{dsc}agent:' .. agentID
  local f = redis.call('hmget', akey, 'state', 'heartbeat', 'capacity', 'dsCount', 'grpcAddr', 'templateReady')
  local state, hb, cap, cnt, grpc, tready = f[1], tonumber(f[2] or '0'), tonumber(f[3] or '0'), tonumber(f[4] or '0'), f[5], f[6]
  if state == 'online' and grpc and grpc ~= '' and tready == '1' then
    if (now - hb) <= hbTimeout then
      local newCnt = redis.call('hincrby', akey, 'dsCount', 1)
      if cap > 0 and newCnt > cap then
        redis.call('hincrby', akey, 'dsCount', -1)
      else
        redis.call('zadd', KEYS[1], newCnt, agentID)
        return {agentID, grpc}
      end
    end
  end
end
return {}
`)
)

type AgentInfo struct {
	AgentID       string
	GrpcAddr      string
	DSCount       int64
	Capacity      int64
	Heartbeat     int64
	State         string
	Runtime       string
	TemplateReady bool
}

type DSInfo struct {
	DSID       string
	RealmID    string
	AgentID    string
	RealmCfgID string
	DSAddr     string
	State      string
	UpdatedAt  int64
}

type RealmInfo struct {
	RealmID    string
	TeamID     string
	RealmCfgID string
	State      string
	DSID       string
	DSAddr     string
	FailReason string
	UpdatedAt  int64
	UserIDList []string
}

func RegisterAgent(ctx context.Context, agentID, grpcAddr string, capacity int64, runtime string, templateReady bool) error {
	now := time.Now().Unix()
	key := AgentKey(agentID)
	// DSA 重连时按仍占槽的 DS 重算 dsCount，避免一律清零导致容量账错。
	dsCount := int64(0)
	if list, err := ListDSByAgent(ctx, agentID); err == nil {
		for _, ds := range list {
			switch ds.State {
			case DSStateCreating, DSStateReady, DSStatePlaying:
				dsCount++
			}
		}
	}
	treadyStr := "0"
	if templateReady {
		treadyStr = "1"
	}
	pipe := store.RedisClient.Pipeline()
	pipe.HSet(ctx, key, map[string]interface{}{
		"agentID":       agentID,
		"grpcAddr":      grpcAddr,
		"capacity":      capacity,
		"dsCount":       dsCount,
		"heartbeat":     now,
		"state":         AgentStateOnline,
		"runtime":       runtime,
		"templateReady": treadyStr,
	})
	pipe.ZAdd(ctx, agentsZSetKey, redis.Z{Score: float64(dsCount), Member: agentID})
	_, err := pipe.Exec(ctx)
	return err
}

func HeartbeatAgent(ctx context.Context, agentID string, dsCount int64, templateReady bool) error {
	now := time.Now().Unix()
	key := AgentKey(agentID)
	exists, err := store.RedisClient.Exists(ctx, key).Result()
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("agent not registered: %s", agentID)
	}
	// dsCount 由选机预占 / 释放维护，心跳不覆盖；但必须 ZAdd 回选机池（超时 ZRem 后否则永远选不上）。
	_ = dsCount
	cur, err := store.RedisClient.HGet(ctx, key, "dsCount").Int64()
	if err != nil && err != redis.Nil {
		return err
	}
	if cur < 0 {
		cur = 0
	}
	treadyStr := "0"
	if templateReady {
		treadyStr = "1"
	}
	pipe := store.RedisClient.Pipeline()
	pipe.HSet(ctx, key, "heartbeat", now, "state", AgentStateOnline, "templateReady", treadyStr)
	pipe.ZAdd(ctx, agentsZSetKey, redis.Z{Score: float64(cur), Member: agentID})
	_, err = pipe.Exec(ctx)
	return err
}

func GetAgent(ctx context.Context, agentID string) (*AgentInfo, error) {
	m, err := store.RedisClient.HGetAll(ctx, AgentKey(agentID)).Result()
	if err != nil {
		return nil, err
	}
	if len(m) == 0 {
		return nil, nil
	}
	return mapToAgent(m), nil
}

// ReserveOptimalAgent 原子选机并 dsCount+1 预占容量；失败需 ReleaseAgentSlot。
func ReserveOptimalAgent(ctx context.Context, heartbeatTimeoutSec int64) (*AgentInfo, error) {
	now := time.Now().Unix()
	res, err := reserveOptimalAgentScript.Run(ctx, store.RedisClient, []string{agentsZSetKey}, now, heartbeatTimeoutSec).Slice()
	if err != nil {
		return nil, err
	}
	if len(res) < 2 {
		return nil, nil
	}
	agentID, _ := res[0].(string)
	grpcAddr, _ := res[1].(string)
	if agentID == "" {
		return nil, nil
	}
	agent, err := GetAgent(ctx, agentID)
	if err != nil {
		return nil, err
	}
	if agent == nil {
		agent = &AgentInfo{AgentID: agentID, GrpcAddr: grpcAddr, State: AgentStateOnline}
	}
	if agent.GrpcAddr == "" {
		agent.GrpcAddr = grpcAddr
	}
	return agent, nil
}

func ReleaseAgentSlot(ctx context.Context, agentID string) error {
	if agentID == "" {
		return nil
	}
	key := AgentKey(agentID)
	n, err := store.RedisClient.HIncrBy(ctx, key, "dsCount", -1).Result()
	if err != nil {
		return err
	}
	if n < 0 {
		_ = store.RedisClient.HSet(ctx, key, "dsCount", 0).Err()
		n = 0
	}
	return store.RedisClient.ZAdd(ctx, agentsZSetKey, redis.Z{Score: float64(n), Member: agentID}).Err()
}

// TryClaimRealmAlloc CAS 抢同一 realm 的分配权，防双开 DS。
// 返回 AllocClaimOK / AllocClaimReady / AllocClaimCreating / AllocClaimMissing；ready 时 dsAddr 非空。
func TryClaimRealmAlloc(ctx context.Context, realmID string) (claim string, dsAddr string, err error) {
	now := strconv.FormatInt(time.Now().Unix(), 10)
	res, err := tryClaimRealmAllocScript.Run(ctx, store.RedisClient,
		[]string{RealmKey(realmID), Realm2DSKey(realmID), realmIndexKey},
		now, realmID).Slice()
	if err != nil {
		return "", "", err
	}
	if len(res) < 1 {
		return AllocClaimCreating, "", nil
	}
	claim, _ = res[0].(string)
	if len(res) > 1 {
		dsAddr, _ = res[1].(string)
	}
	return claim, dsAddr, nil
}

func ReleaseRealmAlloc(ctx context.Context, realmID string) error {
	return store.RedisClient.HDel(ctx, RealmKey(realmID), RealmAllocField).Err()
}

func realmState(ctx context.Context, realmID string) (string, error) {
	state, err := store.RedisClient.HGet(ctx, RealmKey(realmID), "state").Result()
	if err == redis.Nil {
		return "", nil
	}
	return state, err
}

// RealmAllowsAllocate: creating（含刚 claim）才允许继续拉起 DS。
func RealmAllowsAllocate(ctx context.Context, realmID string) (bool, error) {
	state, err := realmState(ctx, realmID)
	if err != nil {
		return false, err
	}
	return state == RealmStateCreating, nil
}

// RealmAllowsReady: creating 可推进 ready；已 ready 幂等；ended/fail 拒绝复活。
func RealmAllowsReady(ctx context.Context, realmID string) (allow bool, alreadyReady bool, err error) {
	state, err := realmState(ctx, realmID)
	if err != nil {
		return false, false, err
	}
	switch state {
	case RealmStateCreating:
		return true, false, nil
	case RealmStateReady, RealmStateInGame:
		return true, true, nil
	default:
		return false, false, nil
	}
}

func MarkAgentOffline(ctx context.Context, agentID string) error {
	pipe := store.RedisClient.Pipeline()
	pipe.HSet(ctx, AgentKey(agentID), "state", AgentStateOffline)
	pipe.ZRem(ctx, agentsZSetKey, agentID)
	_, err := pipe.Exec(ctx)
	return err
}

func ListStaleAgents(ctx context.Context, heartbeatTimeoutSec int64) ([]string, error) {
	now := time.Now().Unix()
	members, err := store.RedisClient.ZRange(ctx, agentsZSetKey, 0, -1).Result()
	if err != nil {
		return nil, err
	}
	var stale []string
	for _, agentID := range members {
		agent, err := GetAgent(ctx, agentID)
		if err != nil || agent == nil {
			continue
		}
		if now-agent.Heartbeat > heartbeatTimeoutSec {
			stale = append(stale, agentID)
		}
	}
	return stale, nil
}

func GenerateDSID() string {
	return primitive.NewObjectID().Hex()
}

// CreateDSMeta 写入 ds 元数据并绑定 realm2ds。
// 调用前须已 ReserveOptimalAgent（容量已预占），此处不再 HIncrBy dsCount。
// 若 GS 已 CleanupRealm，拒绝写入以免复活残片 hash。
// Cluster：{dsc} 与 {realm} 分槽，分两步写；realm 侧失败则回滚 dsc 侧（不 ReleaseAgentSlot，由调用方负责）。
func CreateDSMeta(ctx context.Context, dsID, realmID, agentID, realmCfgID string) error {
	exists, err := store.RedisClient.Exists(ctx, RealmKey(realmID)).Result()
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("realm gone: %s", realmID)
	}
	now := time.Now().Unix()
	dsPipe := store.RedisClient.Pipeline()
	dsPipe.HSet(ctx, DSKey(dsID), map[string]interface{}{
		"dsID":       dsID,
		"realmID":    realmID,
		"agentID":    agentID,
		"realmCfgID": realmCfgID,
		"dsAddr":     "",
		"state":      DSStateCreating,
		"updatedAt":  now,
	})
	dsPipe.SAdd(ctx, dsIndexKey, dsID)
	if _, err = dsPipe.Exec(ctx); err != nil {
		return err
	}

	realmPipe := store.RedisClient.Pipeline()
	realmPipe.HSet(ctx, RealmKey(realmID), "dsID", dsID, "updatedAt", now)
	realmPipe.Set(ctx, Realm2DSKey(realmID), dsID, 0)
	if _, err = realmPipe.Exec(ctx); err != nil {
		_ = store.RedisClient.Del(ctx, DSKey(dsID)).Err()
		_ = store.RedisClient.SRem(ctx, dsIndexKey, dsID).Err()
		return err
	}
	return nil
}

func GetDS(ctx context.Context, dsID string) (*DSInfo, error) {
	m, err := store.RedisClient.HGetAll(ctx, DSKey(dsID)).Result()
	if err != nil {
		return nil, err
	}
	if len(m) == 0 {
		return nil, nil
	}
	return mapToDS(m), nil
}

func GetDSIDByRealm(ctx context.Context, realmID string) (string, error) {
	val, err := store.RedisClient.Get(ctx, Realm2DSKey(realmID)).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

func UpdateDSState(ctx context.Context, dsID, state, dsAddr string) error {
	// Do not recreate a deleted DS hash; sweep/DEAD paths may fire after DeleteDSMeta.
	exists, err := store.RedisClient.Exists(ctx, DSKey(dsID)).Result()
	if err != nil || exists == 0 {
		return err
	}
	now := time.Now().Unix()
	fields := map[string]interface{}{
		"state":     state,
		"updatedAt": now,
	}
	if dsAddr != "" {
		fields["dsAddr"] = dsAddr
	}
	return store.RedisClient.HSet(ctx, DSKey(dsID), fields).Err()
}

func ListDSByAgent(ctx context.Context, agentID string) ([]*DSInfo, error) {
	ids, err := store.RedisClient.SMembers(ctx, dsIndexKey).Result()
	if err != nil {
		return nil, err
	}
	var list []*DSInfo
	for _, id := range ids {
		ds, err := GetDS(ctx, id)
		if err != nil || ds == nil {
			continue
		}
		if ds.AgentID == agentID {
			list = append(list, ds)
		}
	}
	return list, nil
}

func ListCreatingOrPlayingDS(ctx context.Context) ([]*DSInfo, error) {
	ids, err := store.RedisClient.SMembers(ctx, dsIndexKey).Result()
	if err != nil {
		return nil, err
	}
	var list []*DSInfo
	for _, id := range ids {
		ds, err := GetDS(ctx, id)
		if err != nil || ds == nil {
			continue
		}
		if ds.State == DSStateCreating || ds.State == DSStatePlaying || ds.State == DSStateReady {
			list = append(list, ds)
		}
	}
	return list, nil
}

// DeleteDSMeta 删除 DS 元数据并释放选机预占的容量槽。
// 幂等：仅当本次确实删除了 DSKey（Del 返回 1）才 ReleaseAgentSlot，
// 避免 RemoveRealm 与迟到 READY 被拒后的 removeRemoteDS 并发各释放一次，
// 导致 agent dsCount 多减（容量少计、被过度分配）。
// 不 HSet realm 的 dsID=""：若 realm 已被 CleanupRealm 删除，HSet 会重建残片 hash；
// realm2ds 已 Del 足以解绑，realm hash 的 dsID 字段不再有读者依赖。
func DeleteDSMeta(ctx context.Context, dsID string) error {
	ds, err := GetDS(ctx, dsID)
	if err != nil || ds == nil {
		return err
	}
	// {dsc} 与 {realm}2ds 分槽，先解绑再删 DS。
	if ds.RealmID != "" {
		_ = store.RedisClient.Del(ctx, Realm2DSKey(ds.RealmID)).Err()
	}
	pipe := store.RedisClient.Pipeline()
	delCmd := pipe.Del(ctx, DSKey(dsID))
	pipe.SRem(ctx, dsIndexKey, dsID)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return err
	}
	if delCmd.Err() == nil && delCmd.Val() == 1 {
		return ReleaseAgentSlot(ctx, ds.AgentID)
	}
	return nil
}

func WriteRealmSkeleton(ctx context.Context, realmID, teamID, realmCfgID string, userIDs []string) error {
	now := time.Now().Unix()
	key := RealmKey(realmID)
	uidListKey := RealmUIDListKey(realmID)
	pipe := store.RedisClient.Pipeline()
	pipe.HSet(ctx, key, map[string]interface{}{
		"realmID":    realmID,
		"teamID":     teamID,
		"realmCfgID": realmCfgID,
		"state":      RealmStateCreating,
		"dsID":       "",
		"dsAddr":     "",
		"failReason": "",
		"updatedAt":  now,
	})
	pipe.Del(ctx, uidListKey)
	for _, uid := range userIDs {
		pipe.RPush(ctx, uidListKey, uid)
		pipe.Set(ctx, RealmUserKey(uid), realmID, 0)
	}
	pipe.SAdd(ctx, realmIndexKey, realmID)
	_, err := pipe.Exec(ctx)
	return err
}

func GetRealmIDByUser(ctx context.Context, userID string) (string, error) {
	val, err := store.RedisClient.Get(ctx, RealmUserKey(userID)).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

func GetRealm(ctx context.Context, realmID string) (*RealmInfo, error) {
	m, err := store.RedisClient.HGetAll(ctx, RealmKey(realmID)).Result()
	if err != nil {
		return nil, err
	}
	if len(m) == 0 {
		return nil, nil
	}
	info := mapToRealm(m)
	uids, err := store.RedisClient.LRange(ctx, RealmUIDListKey(realmID), 0, -1).Result()
	if err != nil {
		return nil, err
	}
	info.UserIDList = uids
	return info, nil
}

func MarkRealmReady(ctx context.Context, realmID, dsID, dsAddr string) error {
	exists, err := store.RedisClient.Exists(ctx, RealmKey(realmID)).Result()
	if err != nil || exists == 0 {
		return err
	}
	uids, _ := store.RedisClient.LRange(ctx, RealmUIDListKey(realmID), 0, -1).Result()
	teamID, _ := store.RedisClient.HGet(ctx, RealmKey(realmID), "teamID").Result()
	realmCfgID, _ := store.RedisClient.HGet(ctx, RealmKey(realmID), "realmCfgID").Result()
	payload, _ := json.Marshal(RealmPubPayload{
		DsAddr: dsAddr, TeamID: teamID, RealmCfgID: realmCfgID, UserIDs: uids,
	})
	now := time.Now().Unix()
	key := RealmKey(realmID)
	pipe := store.RedisClient.Pipeline()
	pipe.HSet(ctx, key, map[string]interface{}{
		"state":     RealmStateReady,
		"dsID":      dsID,
		"dsAddr":    dsAddr,
		"updatedAt": now,
	})
	pipe.HDel(ctx, key, RealmAllocField)
	if _, err = pipe.Exec(ctx); err != nil {
		return err
	}
	// PUBLISH 频道无 {realm} 槽标签，Cluster 下不可与 realm KEY 同 pipeline。
	return store.RedisClient.Publish(ctx, RealmReadyChan(realmID), string(payload)).Err()
}

func MarkRealmFail(ctx context.Context, realmID, reason string) error {
	exists, err := store.RedisClient.Exists(ctx, RealmKey(realmID)).Result()
	if err != nil || exists == 0 {
		return err // already cleaned by GS; do not recreate stub
	}
	uids, _ := store.RedisClient.LRange(ctx, RealmUIDListKey(realmID), 0, -1).Result()
	payload, _ := json.Marshal(RealmPubPayload{Reason: reason, UserIDs: uids})
	now := time.Now().Unix()
	key := RealmKey(realmID)
	pipe := store.RedisClient.Pipeline()
	pipe.HSet(ctx, key, map[string]interface{}{
		"state":      RealmStateFail,
		"failReason": reason,
		"updatedAt":  now,
	})
	pipe.HDel(ctx, key, RealmAllocField)
	if _, err = pipe.Exec(ctx); err != nil {
		return err
	}
	return store.RedisClient.Publish(ctx, RealmFailChan(realmID), string(payload)).Err()
}

func MarkRealmInGame(ctx context.Context, realmID string) error {
	exists, err := store.RedisClient.Exists(ctx, RealmKey(realmID)).Result()
	if err != nil || exists == 0 {
		return err
	}
	return store.RedisClient.HSet(ctx, RealmKey(realmID), "state", RealmStateInGame, "updatedAt", time.Now().Unix()).Err()
}

func MarkRealmEnded(ctx context.Context, realmID string) error {
	exists, err := store.RedisClient.Exists(ctx, RealmKey(realmID)).Result()
	if err != nil || exists == 0 {
		return err
	}
	return store.RedisClient.HSet(ctx, RealmKey(realmID), "state", RealmStateEnded, "updatedAt", time.Now().Unix()).Err()
}

func ClearRealmUserMappings(ctx context.Context, realmID string) error {
	uids, err := store.RedisClient.LRange(ctx, RealmUIDListKey(realmID), 0, -1).Result()
	if err != nil {
		return err
	}
	pipe := store.RedisClient.Pipeline()
	for _, uid := range uids {
		pipe.Del(ctx, RealmUserKey(uid))
	}
	_, err = pipe.Exec(ctx)
	return err
}

// RemoveUserFromRealm removes a single user from a realm's uid list and mapping.
// Returns the remaining member count. If 0, realm is empty.
func RemoveUserFromRealm(ctx context.Context, realmID, userID string) (int64, error) {
	pipe := store.RedisClient.Pipeline()
	pipe.LRem(ctx, RealmUIDListKey(realmID), 0, userID)
	pipe.Del(ctx, RealmUserKey(userID))
	res, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return 0, err
	}
	removed := int64(0)
	if len(res) > 0 {
		if cmd, ok := res[0].(*redis.IntCmd); ok {
			removed = cmd.Val()
		}
	}
	_ = removed
	remaining, err := store.RedisClient.LLen(ctx, RealmUIDListKey(realmID)).Result()
	return remaining, err
}

// CleanupRealm deletes all realm keys and user mappings (used after ended/fail).
func CleanupRealm(ctx context.Context, realmID string) error {
	uids, err := store.RedisClient.LRange(ctx, RealmUIDListKey(realmID), 0, -1).Result()
	if err != nil {
		return err
	}
	pipe := store.RedisClient.Pipeline()
	for _, uid := range uids {
		pipe.Del(ctx, RealmUserKey(uid))
	}
	pipe.Del(ctx, RealmKey(realmID))
	pipe.Del(ctx, RealmUIDListKey(realmID))
	pipe.Del(ctx, Realm2DSKey(realmID))
	pipe.SRem(ctx, realmIndexKey, realmID)
	_, err = pipe.Exec(ctx)
	return err
}

// ListRealmIDs returns realm IDs tracked in the index (for orphan sweep).
func ListRealmIDs(ctx context.Context) ([]string, error) {
	return store.RedisClient.SMembers(ctx, realmIndexKey).Result()
}

// SweepRealmFragments removes stub/fail/ended realms past timeout, and creating
// realms that never got a live DS (network-partition leftovers).
func SweepRealmFragments(ctx context.Context, createTimeoutSec, staleTimeoutSec int64) (cleaned int, err error) {
	ids, err := ListRealmIDs(ctx)
	if err != nil {
		return 0, err
	}
	now := time.Now().Unix()
	for _, realmID := range ids {
		info, err := GetRealm(ctx, realmID)
		if err != nil {
			continue
		}
		if info == nil {
			_ = store.RedisClient.SRem(ctx, realmIndexKey, realmID).Err()
			cleaned++
			continue
		}
		age := now - info.UpdatedAt
		if info.UpdatedAt <= 0 {
			age = staleTimeoutSec + 1 // unknown age: treat as stale once past threshold
		}
		switch info.State {
		case "", RealmStateFail, RealmStateEnded:
			if age > staleTimeoutSec {
				_ = CleanupRealm(ctx, realmID)
				cleaned++
			}
		case RealmStateCreating:
			if age > createTimeoutSec {
				dsID := info.DSID
				if dsID == "" {
					dsID, _ = GetDSIDByRealm(ctx, realmID)
				}
				if dsID != "" {
					ds, _ := GetDS(ctx, dsID)
					if ds != nil && (ds.State == DSStateCreating || ds.State == DSStateReady || ds.State == DSStatePlaying) {
						continue // live DS path handled by DS sweep
					}
				}
				_ = CleanupRealm(ctx, realmID)
				cleaned++
			}
		case RealmStateReady, RealmStateInGame:
			// Orphan ready/ingame with no live DS (e.g. MarkRealmReady raced with
			// CleanupRealm and left a stub, or DS died without reporting). Reclaim
			// past the ready idle window so players are not stuck on a dead realm.
			if age > staleTimeoutSec {
				dsID := info.DSID
				if dsID == "" {
					dsID, _ = GetDSIDByRealm(ctx, realmID)
				}
				live := false
				if dsID != "" {
					if ds, _ := GetDS(ctx, dsID); ds != nil &&
						(ds.State == DSStateReady || ds.State == DSStatePlaying) {
						live = true
					}
				}
				if !live {
					_ = CleanupRealm(ctx, realmID)
					cleaned++
				}
			}
		}
	}
	return cleaned, nil
}

func mapToAgent(m map[string]string) *AgentInfo {
	dsCount, _ := strconv.ParseInt(m["dsCount"], 10, 64)
	capacity, _ := strconv.ParseInt(m["capacity"], 10, 64)
	hb, _ := strconv.ParseInt(m["heartbeat"], 10, 64)
	return &AgentInfo{
		AgentID:       m["agentID"],
		GrpcAddr:      m["grpcAddr"],
		DSCount:       dsCount,
		Capacity:      capacity,
		Heartbeat:     hb,
		State:         m["state"],
		Runtime:       m["runtime"],
		TemplateReady: m["templateReady"] == "1",
	}
}

func mapToDS(m map[string]string) *DSInfo {
	updated, _ := strconv.ParseInt(m["updatedAt"], 10, 64)
	return &DSInfo{
		DSID:       m["dsID"],
		RealmID:    m["realmID"],
		AgentID:    m["agentID"],
		RealmCfgID: m["realmCfgID"],
		DSAddr:     m["dsAddr"],
		State:      m["state"],
		UpdatedAt:  updated,
	}
}

func mapToRealm(m map[string]string) *RealmInfo {
	updated, _ := strconv.ParseInt(m["updatedAt"], 10, 64)
	return &RealmInfo{
		RealmID:    m["realmID"],
		TeamID:     m["teamID"],
		RealmCfgID: m["realmCfgID"],
		State:      m["state"],
		DSID:       m["dsID"],
		DSAddr:     m["dsAddr"],
		FailReason: m["failReason"],
		UpdatedAt:  updated,
	}
}

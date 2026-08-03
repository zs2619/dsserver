package dsastore

const (
	RealmStateCreating = "creating"
	RealmStateReady    = "ready"
	RealmStateInGame   = "ingame"
	RealmStateFail     = "fail"
	RealmStateEnded    = "ended"

	DSStateCreating = "creating"
	DSStateReady    = "ready"
	DSStatePlaying  = "playing"
	DSStateEnded    = "ended"
	DSStateDead     = "dead"

	AgentStateOnline  = "online"
	AgentStateOffline = "offline"

	// Hash tags (Redis Cluster):
	//   {realm} — realm hash / uidList / user→realm / realm2ds / realmIndex
	//             同槽，供 TryClaim / WriteSkeleton / Cleanup 原子脚本与 pipeline。
	//   {dsc}   — agent / ds meta / dsIndex（选机与 DS 生命周期）。
	// 跨槽操作（如 CreateDSMeta）必须拆成两次命令，不可混 KEYS / 同 pipeline。
	realmKeyPrefix     = "{realm}_"
	realmUidListSuffix = "_uidList"
	realmUserKeyPrefix = "{realm}u:"
	realm2DSKeyPref    = "{realm}2ds:"
	realmIndexKey      = "{realm}Index"
	realmReadyChannel  = "realm:ready:"
	realmFailChannel   = "realm:fail:"

	dsKeyPrefix    = "{dsc}ds:"
	agentKeyPrefix = "{dsc}agent:"
	agentsZSetKey  = "{dsc}agents"
	dsIndexKey     = "{dsc}dsIndex"

	// RealmAllocField: HSETNX 占位，防同一 realm 并发分配双开 DS
	RealmAllocField = "alloc"
)

// AllocClaim results from TryClaimRealmAlloc.
const (
	AllocClaimOK       = "ok"
	AllocClaimReady    = "ready"
	AllocClaimCreating = "creating"
	AllocClaimMissing  = "missing" // GS skeleton gone; do not recreate stub
)

func RealmKey(realmID string) string {
	return realmKeyPrefix + realmID
}

func RealmUIDListKey(realmID string) string {
	return RealmKey(realmID) + realmUidListSuffix
}

func RealmUserKey(userID string) string {
	return realmUserKeyPrefix + userID
}

func RealmReadyChan(realmID string) string {
	return realmReadyChannel + realmID
}

func RealmFailChan(realmID string) string {
	return realmFailChannel + realmID
}

func DSKey(dsID string) string {
	return dsKeyPrefix + dsID
}

func Realm2DSKey(realmID string) string {
	return realm2DSKeyPref + realmID
}

func AgentKey(agentID string) string {
	return agentKeyPrefix + agentID
}

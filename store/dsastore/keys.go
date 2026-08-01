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

	realmKeyPrefix     = "{realm}_"
	realmUidListSuffix = "_uidList"
	realmUserKeyPrefix = "{realm}u:"
	realmReadyChannel  = "realm:ready:"
	realmFailChannel   = "realm:fail:"

	dsKeyPrefix     = "{dsc}ds:"
	realm2DSKeyPref = "{dsc}realm2ds:"
	agentKeyPrefix  = "{dsc}agent:"
	agentsZSetKey   = "{dsc}agents"
	dsIndexKey      = "{dsc}dsIndex"
	realmIndexKey   = "{dsc}realmIndex"

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

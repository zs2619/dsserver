package team

import (
	"context"
	"dsservices/common"
	"dsservices/config"
	"dsservices/gameserver/realm"
	"dsservices/store"
	"dsservices/store/dsastore"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TeamInfo struct {
	TeamID     string   `json:"teamID"`
	RealmCfgID string   `json:"realmCfgID"`
	UserIDList []string `json:"userIDList"`
	CreateTime int64    `json:"createTime"`
	Timeout    int64    `json:"timeout"`
	State      int64    `json:"state"` // 0 matching 1 closed
	ForceStart int64    `json:"forceStart"`
}

const (
	teamExpireTime     = 3600 * 12
	teamZsetKeyName    = "{team}Zset"
	teamHashKeyLuaName = "{team}"
	teamUidListPostfix = "_uidList"
	teamUserKeyPrefix  = "{team}u:"
	teamMaxMembers     = 5

	teamStateMatching = 0
	teamStateClosed   = 1
)

var (
	quickJoinRealmTeamScript *redis.Script
	updateTimeoutTeamScript  *redis.Script
	quitTeamScript           *redis.Script
	convertTeamScript        *redis.Script
	dissolveTeamScript       *redis.Script
)

func Init() error {
	err := InitScript()
	if err != nil {
		return err
	}
	common.GCron.AddFunc("@every 1s", UpdateTimeoutTeam)
	return nil
}

func InitScript() error {
	quickJoinRealmTeamStr := fmt.Sprintf(`
	local uid ,nowTime,timeout,teamID,teamMax,realmCfgID= ARGV[1],tonumber(ARGV[2]),tonumber(ARGV[3]),ARGV[4],tonumber(ARGV[5]),ARGV[6]
	local userKey='%s'..uid
	local existTeam=redis.call('get',userKey)
	if existTeam
	then
		return existTeam
	end
	local realmUser=redis.call('get','{realm}u:'..uid)
	if realmUser
	then
		return redis.error_reply('already in realm')
	end
	local retSet = redis.call('zrange',KEYS[1],'0','-1')
	for key,value in ipairs(retSet)
	do
		local keyUidList=value..'%s'
		local num=redis.call('llen',keyUidList)
		if num< teamMax
		then
			local teamCfg=redis.call('hget',value,'realmCfgID')
			if teamCfg ~= realmCfgID
			then
				-- skip teams of other realmCfgID
				goto continue
			end
			local ul=redis.call('lrange', keyUidList,0,-1)
			for key1,value1 in ipairs(ul)
			do
				if uid==value1
				then
					redis.call('set',userKey,value)
					return value
				end
			end
			redis.call('rpush',keyUidList,uid)
			redis.call('set',userKey,value)
			num=num+1
			if num>= teamMax
			then
				redis.call('hmset',value,'timeout',0)
				redis.call('zadd',KEYS[1],0,value)
			end
			return value
		else
			redis.call('hmset',value,'timeout',0)
			redis.call('zadd',KEYS[1],0,value)
		end
		::continue::
	end

	local newTeamKey='%s'..'_'..teamID
	local keyUidList=newTeamKey..'%s'
	redis.call('hmset',newTeamKey,'teamID',teamID,'createTime',ARGV[2],'timeout',ARGV[3],'state',0,'realmCfgID',realmCfgID,'forceStart',0)
	redis.call('EXPIRE',newTeamKey,%d)
	redis.call('rpush',keyUidList,uid)
	redis.call('EXPIRE',keyUidList,%d)
	redis.call('zadd',KEYS[1],timeout,newTeamKey)
	redis.call('set',userKey,newTeamKey)
	return newTeamKey`, teamUserKeyPrefix, teamUidListPostfix, teamHashKeyLuaName, teamUidListPostfix, teamExpireTime, teamExpireTime)

	quickJoinRealmTeamScript = redis.NewScript(quickJoinRealmTeamStr)

	quitTeamStr := fmt.Sprintf(`
	local userID= ARGV[1]
	local teamIDKey= KEYS[2]
	local keyUidList=teamIDKey..'%s'
	local userKey='%s'..userID
	redis.call('lrem',keyUidList,0,userID)
	redis.call('del',userKey)

	local num=redis.call('llen',keyUidList)
	if num==0
	then
		redis.call('del',teamIDKey)
		redis.call('del',keyUidList)
		redis.call('zrem',KEYS[1],teamIDKey)
		return 1
	end
	return 0`, teamUidListPostfix, teamUserKeyPrefix)
	quitTeamScript = redis.NewScript(quitTeamStr)

	updateTimeoutTeamStr := fmt.Sprintf(`
	local nowTime= tonumber(ARGV[1])
	local ret={}
	local retSet = redis.call('zrange',KEYS[1],'0','3')
	for key,value in ipairs(retSet)
	do
		local score=redis.call('zscore',KEYS[1],value)
		if tonumber(score)<= nowTime
		then
			local keyUidList=value..'%s'
			local ul=redis.call('lrange', keyUidList,0,-1)
			local fields=redis.call('hmget',value,'realmCfgID','forceStart','state','teamID')
			ret[value]={realmCfgID=fields[1],forceStart=fields[2],state=fields[3],teamID=fields[4],userIDList=ul}
			break
		end
	end
	return cjson.encode(ret)
	`, teamUidListPostfix)
	updateTimeoutTeamScript = redis.NewScript(updateTimeoutTeamStr)

	// CAS matching→closed, remove from zset, clear team user keys; return ok
	convertTeamStr := fmt.Sprintf(`
	local teamKey=KEYS[1]
	local state=redis.call('hget',teamKey,'state')
	if not state or tonumber(state)~=0
	then
		return 0
	end
	redis.call('hmset',teamKey,'state',1)
	redis.call('zrem','%s',teamKey)
	local keyUidList=teamKey..'%s'
	local ul=redis.call('lrange', keyUidList,0,-1)
	for i,uid in ipairs(ul)
	do
		redis.call('del','%s'..uid)
	end
	redis.call('del',teamKey)
	redis.call('del',keyUidList)
	return 1
	`, teamZsetKeyName, teamUidListPostfix, teamUserKeyPrefix)
	convertTeamScript = redis.NewScript(convertTeamStr)

	dissolveTeamStr := fmt.Sprintf(`
	local teamKey=KEYS[1]
	local state=redis.call('hget',teamKey,'state')
	if not state or tonumber(state)~=0
	then
		return 0
	end
	local keyUidList=teamKey..'%s'
	local ul=redis.call('lrange', keyUidList,0,-1)
	for i,uid in ipairs(ul)
	do
		redis.call('del','%s'..uid)
	end
	redis.call('del',teamKey)
	redis.call('del',keyUidList)
	redis.call('zrem','%s',teamKey)
	return 1
	`, teamUidListPostfix, teamUserKeyPrefix, teamZsetKeyName)
	dissolveTeamScript = redis.NewScript(dissolveTeamStr)
	return nil
}

func QuickJoinRealmTeam(userID, realmCfgID string) (teamID string, err error) {
	matchTimeout := config.GameConfig.Team.MatchTimeoutMs
	if matchTimeout <= 0 {
		matchTimeout = 10000
	}
	argv1 := userID
	argv2 := common.GetNowTime().UnixMilli()
	argv3 := common.GetNowTime().UnixMilli() + matchTimeout
	argv4 := primitive.NewObjectID().Hex()
	argv5 := teamMaxMembers
	argv6 := realmCfgID

	cmd := quickJoinRealmTeamScript.Run(context.TODO(), store.RedisClient, []string{teamZsetKeyName}, argv1, argv2, argv3, argv4, argv5, argv6)
	result, err := cmd.Result()
	if err != nil {
		return "", err
	}
	teamID = result.(string)
	return
}

func StartRealm(userID, teamID string) error {
	pipe := store.RedisClient.Pipeline()
	pipe.HSet(context.TODO(), teamID, "forceStart", 1, "timeout", 0)
	pipe.ZAdd(context.TODO(), teamZsetKeyName, redis.Z{Score: 0, Member: teamID})
	_, err := pipe.Exec(context.TODO())
	return err
}

func GetRealmTeam(teamID string) (teamInfo *TeamInfo, err error) {
	userListKey := teamID + teamUidListPostfix
	cmderList, err := store.RedisClient.Pipelined(context.TODO(), func(pipe redis.Pipeliner) error {
		pipe.HGetAll(context.TODO(), teamID).Result()
		pipe.LRange(context.TODO(), userListKey, 0, -1).Result()
		return nil
	})

	teamInfo = new(TeamInfo)
	teamInfoCmd, ok := cmderList[0].(*redis.MapStringStringCmd)
	if ok {
		var teamInfoMap map[string]string
		teamInfoMap, err = teamInfoCmd.Result()
		if err != nil {
			return
		}
		if len(teamInfoMap) == 0 {
			err = fmt.Errorf("teamID nil")
			teamInfo = nil
			return
		}
		teamInfo.TeamID = teamHashKeyLuaName + "_" + teamInfoMap["teamID"]
		teamInfo.RealmCfgID = teamInfoMap["realmCfgID"]
		createTime, _ := strconv.ParseInt(teamInfoMap["createTime"], 10, 64)
		teamInfo.CreateTime = createTime
		timeout, _ := strconv.ParseInt(teamInfoMap["timeout"], 10, 64)
		teamInfo.Timeout = timeout
		state, _ := strconv.ParseInt(teamInfoMap["state"], 10, 64)
		teamInfo.State = state
		force, _ := strconv.ParseInt(teamInfoMap["forceStart"], 10, 64)
		teamInfo.ForceStart = force
	}

	userIDListCmd, ok := cmderList[1].(*redis.StringSliceCmd)
	if ok {
		var userIDList []string
		userIDList, err = userIDListCmd.Result()
		if err != nil {
			teamInfo = nil
			return
		}
		teamInfo.UserIDList = userIDList
	}
	return
}

func QuitRealmTeam(userID, teamID string) (int, error) {
	argv1 := userID
	cmd := quitTeamScript.Run(context.TODO(), store.RedisClient, []string{teamZsetKeyName, teamID}, argv1)
	result, err := cmd.Int()
	if err != nil {
		return 0, err
	}
	return result, nil
}

// QuitRealm removes a user from an existing realm (mid-game leave).
// Returns remaining member count. Caller should RemoveRealm when 0.
func QuitRealm(userID, realmID string) (int64, error) {
	remaining, err := dsastore.RemoveUserFromRealm(context.TODO(), realmID, userID)
	if err != nil {
		return 0, err
	}
	if remaining <= 0 {
		realm.RemoveRealm(realmID)
	}
	return remaining, nil
}

func UpdateTimeoutTeam() {
	teamUserMap, err := GetTimeoutTeam()
	if err != nil {
		logrus.WithError(err).Warn("GetTimeoutTeam")
		return
	}
	for teamKey, info := range teamUserMap {
		force := info.ForceStart == 1
		full := len(info.UserIDList) >= teamMaxMembers
		if force || full {
			convertTeamToRealm(teamKey, info)
		} else {
			dissolveTeam(teamKey)
			logrus.WithField("teamKey", teamKey).Info("team match timeout dissolved")
		}
	}
}

func convertTeamToRealm(teamKey string, info *TeamInfo) {
	// CAS 先关 team，避免多 GS 并发各写一套 realm 骨架互相覆盖 user 映射。
	n, err := convertTeamScript.Run(context.TODO(), store.RedisClient, []string{teamKey}).Int()
	if err != nil || n != 1 {
		return
	}
	realmID := primitive.NewObjectID().Hex()
	teamID := info.TeamID
	if teamID == "" {
		teamID = teamKey
	}
	ctx := context.TODO()
	err = dsastore.WriteRealmSkeleton(ctx, realmID, teamID, info.RealmCfgID, info.UserIDList)
	if err != nil {
		logrus.WithError(err).Error("WriteRealmSkeleton after team convert")
		// skeleton 可能部分写入（pipeline 非原子）：清残片 + 显式删 user 映射，
		// 否则 {realm}u:<uid> 残留会阻止玩家重新 QuickJoin。
		_ = dsastore.CleanupRealm(ctx, realmID)
		pipe := store.RedisClient.Pipeline()
		for _, uid := range info.UserIDList {
			pipe.Del(ctx, dsastore.RealmUserKey(uid))
		}
		_, _ = pipe.Exec(ctx)
		if realm.PushRealmFailToLocalPlayers != nil {
			realm.PushRealmFailToLocalPlayers(info.UserIDList, realmID)
		}
		return
	}
	logrus.WithFields(logrus.Fields{
		"teamKey": teamKey, "realmID": realmID, "users": len(info.UserIDList),
	}).Info("team converted to realm")

	go realm.CreateRealmAndWatch(realmID, teamID, info.RealmCfgID, info.UserIDList)
}

func dissolveTeam(teamKey string) {
	_, _ = dissolveTeamScript.Run(context.TODO(), store.RedisClient, []string{teamKey}).Int()
}

func GetTimeoutTeam() (teamUserMap map[string]*TeamInfo, err error) {
	teamUserMap = make(map[string]*TeamInfo)
	argv1 := common.GetNowTime().UnixMilli()
	cmd := updateTimeoutTeamScript.Run(context.TODO(), store.RedisClient, []string{teamZsetKeyName}, argv1)
	result, err := cmd.Result()
	if err != nil {
		return
	}
	retStr := result.(string)
	type rawTeam struct {
		RealmCfgID string   `json:"realmCfgID"`
		ForceStart string   `json:"forceStart"`
		State      string   `json:"state"`
		TeamID     string   `json:"teamID"`
		UserIDList []string `json:"userIDList"`
	}
	raw := make(map[string]*rawTeam)
	err = json.Unmarshal([]byte(retStr), &raw)
	if err != nil {
		return
	}
	for k, v := range raw {
		force, _ := strconv.ParseInt(v.ForceStart, 10, 64)
		state, _ := strconv.ParseInt(v.State, 10, 64)
		teamUserMap[k] = &TeamInfo{
			TeamID:     teamHashKeyLuaName + "_" + v.TeamID,
			RealmCfgID: v.RealmCfgID,
			UserIDList: v.UserIDList,
			ForceStart: force,
			State:      state,
		}
	}
	return
}

func GetTeamUserList(teamID string) (userList []string, err error) {
	userListKey := teamID + teamUidListPostfix
	userList, err = store.RedisClient.LRange(context.TODO(), userListKey, 0, -1).Result()
	return
}

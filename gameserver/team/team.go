package team

import (
	"context"
	"dsservices/common"
	"dsservices/store"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TeamInfo struct {
	TeamID     string   `json:"teamID"`
	RealmCfgID string   `json:"realmCfgID"`
	UserIDList []string `json:"userIDList"`
	CreateTime int64    `json:"createTime"`
	Timeout    int64    `json:"timeout"`
	State      int64    `json:"state"` ///0 初始 1匹配结束 2realm准备好通知玩家 3游戏开始 4游戏结束
	DSAddr     string   `json:"dsAddr"`
}

const (
	teamExpireTime     = 3600 * 12
	teamZsetKeyName    = "{team}Zset"
	teamHashKeyLuaName = "{team}"
	teamUidListPostfix = "_uidList"
)

var (
	quickJoinRealmTeamStr    string
	quickJoinRealmTeamScript *redis.Script

	updateTimeoutTeamStr    string
	updateTimeoutTeamScript *redis.Script

	quitTeamStr    string
	quitTeamScript *redis.Script
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
	quickJoinRealmTeamStr = fmt.Sprintf(`
	local uid ,nowTime,timeout,teamID,teamMax,realmCfgID= ARGV[1],tonumber(ARGV[2]),tonumber(ARGV[3]),ARGV[4],tonumber(ARGV[5]),ARGV[6]
	local retSet = redis.call('zrange',KEYS[1],'0','-1')
	for key,value in ipairs(retSet)
	do
		local keyUidList=value..'%s'
		local num=redis.call('llen',keyUidList)
		if num< teamMax
		then
			local ul=redis.call('lrange', keyUidList,0,-1)
			for key1,value1 in ipairs(ul)
			do
				if uid==value1
				then
					return value 
				end
			end
			redis.call('rpush',keyUidList,uid)
			return value
		else
			redis.call('hmset',value,'timeout',0)
			redis.call('zadd',KEYS[1],0,value)
		end
	end

	local newTeamKey='%s'..'_'..teamID
	local keyUidList=newTeamKey..'%s'
	redis.call('hmset',newTeamKey,'teamID',teamID,'createTime',ARGV[2],'timeout',ARGV[3],'state',0,'realmCfgID',realmCfgID)
	redis.call('EXPIRE',newTeamKey,%d)
	redis.call('rpush',keyUidList,uid)
	redis.call('EXPIRE',keyUidList,%d)
	redis.call('zadd',KEYS[1],timeout,newTeamKey)

	return newTeamKey`, teamUidListPostfix, teamHashKeyLuaName, teamUidListPostfix, teamExpireTime, teamExpireTime)

	quickJoinRealmTeamScript = redis.NewScript(quickJoinRealmTeamStr)

	quitTeamStr = fmt.Sprintf(`
	local userID= ARGV[1]
	local teamIDKey= KEYS[2]
	local keyUidList=teamIDKey..'%s'
	redis.call('lrem',keyUidList,0,userID)

	local num=redis.call('llen',keyUidList)
	if num==0 
	then
		redis.call('del',teamIDKey)
		redis.call('del',keyUidList)
		redis.call('zrem',KEYS[1],teamIDKey)
		return 1
	end
	return 0`, teamUidListPostfix)
	quitTeamScript = redis.NewScript(quitTeamStr)

	updateTimeoutTeamStr = fmt.Sprintf(`
	local nowTime= tonumber(ARGV[1])
	local ret={}
	local retSet = redis.call('zrange',KEYS[1],'0','3')
	for key,value in ipairs(retSet)
	do
		local score=redis.call('zscore',KEYS[1],value)
		if tonumber(score)<= nowTime
		then
			local keyUidList=value..'%s'
			redis.call('hmset',value,'state',1)
			redis.call('zrem',KEYS[1],value)
			local ul=redis.call('lrange', keyUidList,0,-1)
			local id=redis.call('hmget',value,'realmCfgID')
			ret[value]={realmCfgID=id[1],userIDList=ul}
			break
		end
	end 
	return cjson.encode(ret)
	`, teamUidListPostfix)
	updateTimeoutTeamScript = redis.NewScript(updateTimeoutTeamStr)
	return nil
}

//hash set 组队信息
//zset 超时排序

func QuickJoinRealmTeam(userID, realmCfgID string) (teamID string, err error) {
	argv1 := userID
	argv2 := common.GetNowTime().UnixMilli()
	argv3 := common.GetNowTime().UnixMilli() + 10*1000
	argv4 := primitive.NewObjectID().Hex()
	argv5 := 5
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
	err := store.RedisClient.ZAdd(context.TODO(), teamZsetKeyName, redis.Z{Score: 0, Member: teamID}).Err()
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
		teamInfo.DSAddr = teamInfoMap["dsAddr"]
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

func UpdateTimeoutTeam() {
	_, err := GetTimeoutTeam()
	if err != nil {
		//TODO
		return
	}
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
	err = json.Unmarshal([]byte(retStr), &teamUserMap)
	return

}
func GetTeamUserList(teamID string) (userList []string, err error) {
	userListKey := teamID + teamUidListPostfix
	userList, err = store.RedisClient.LRange(context.TODO(), userListKey, 0, -1).Result()
	return
}

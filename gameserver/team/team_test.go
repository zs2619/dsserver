package team_test

import (
	"dsservices/common"
	"dsservices/gameserver/team"
	"dsservices/test"
	"testing"

	"github.com/stretchr/testify/suite"
)

type TeamtilTestSuite struct {
	test.BaseTestSuite
}

func TestTeamessUtilTest(t *testing.T) {
	team.InitScript()
	suite.Run(t, new(TeamtilTestSuite))
}

func (t *TeamtilTestSuite) TestTeam() {
	userID := "userID1"
	realmCfgID := "realmCfgID"
	teamID, err := team.QuickJoinRealmTeam(userID, realmCfgID)
	t.Nil(err)
	t.NotEmpty(teamID)

	teamInfo, err := team.GetRealmTeam(teamID + "ss")
	t.NotNil(err)
	t.Nil(teamInfo)

	teamInfo, err = team.GetRealmTeam(teamID)
	t.Nil(err)
	t.Equal(teamInfo.TeamID, teamID)
	t.Equal(teamInfo.RealmCfgID, realmCfgID)
	t.Equal(teamInfo.UserIDList[0], userID)

	userID2 := "userID2"
	teamID, err = team.QuickJoinRealmTeam(userID2, realmCfgID)
	t.Nil(err)
	t.NotEmpty(teamID)

	teamInfo, err = team.GetRealmTeam(teamID)
	t.Nil(err)
	t.Equal(teamInfo.TeamID, teamID)
	t.Equal(teamInfo.RealmCfgID, realmCfgID)
	t.Equal(teamInfo.UserIDList[0], userID)
	t.Equal(teamInfo.UserIDList[1], userID2)

	ret, err := team.QuitRealmTeam(userID, teamID)
	t.Nil(err)
	t.Equal(ret, 0)

	teamInfo, err = team.GetRealmTeam(teamID)
	t.Nil(err)
	t.Equal(teamInfo.TeamID, teamID)
	t.Equal(teamInfo.RealmCfgID, realmCfgID)
	t.Equal(teamInfo.UserIDList[0], userID2)

	ret, err = team.QuitRealmTeam(userID2, teamID)
	t.Nil(err)
	t.Equal(ret, 1)
	teamInfo, err = team.GetRealmTeam(teamID)
	t.NotNil(err)
	t.Nil(teamInfo)

	common.OffSetTime = 120
	team.UpdateTimeoutTeam()
}

func (t *TeamtilTestSuite) TestTeamStart() {
	userID := "userID1"
	realmCfgID := "realmCfgID"
	teamID, err := team.QuickJoinRealmTeam(userID, realmCfgID)
	t.Nil(err)
	t.NotEmpty(teamID)

	err = team.StartRealm(userID, teamID)
	t.Nil(err)
	teamUserMap, err := team.GetTimeoutTeam()
	t.Nil(err)
	t.Equal(len(teamUserMap), 1)

	teamInfo, err := team.GetRealmTeam(teamID)
	t.Nil(err)
	t.Equal(teamInfo.TeamID, teamID)
	t.Equal(teamInfo.RealmCfgID, realmCfgID)
	t.Equal(teamInfo.State, int64(1))
	t.Equal(teamInfo.UserIDList[0], userID)
}

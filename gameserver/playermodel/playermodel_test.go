package playermodel_test

import (
	"dsservices/gameserver/playermodel"
	"dsservices/gameserver/user"
	"dsservices/test"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PlayerUtilTestSuite struct {
	test.BaseTestSuite
}

func TestPlayerUtilTest(t *testing.T) {
	suite.Run(t, new(PlayerUtilTestSuite))
}

func (t *PlayerUtilTestSuite) TestPlayer() {
	userID := "userID"
	uidObj := primitive.NewObjectID()
	ip := "ip"

	pd, err := playermodel.CreatePlayerModel(userID, uidObj, ip)
	t.Nil(err)
	t.NotNil(pd)
	t.Equal(pd.UserID, userID)
	t.Equal(pd.IP, ip)
	t.Equal(pd.PlayerID.Hex(), uidObj.Hex())

	userID1 := "userID1"
	ip1 := "ip1"
	platform1 := "local"

	err = playermodel.CreateUserPlayer(userID1, platform1, ip1)
	t.Nil(err)
	userRet, err := user.GetUser(userID1)
	t.Nil(err)
	pm, err := playermodel.LoadPlayerModel(userRet.PlayerID.Hex())
	t.Nil(err)
	t.NotNil(pm)
	t.Equal(pm.UserID, userID1)
}

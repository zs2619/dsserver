package user_test

import (
	"dsservices/gameserver/user"
	"dsservices/test"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserUtilTestSuite struct {
	test.BaseTestSuite
}

func TestUserUtilTest(t *testing.T) {
	suite.Run(t, new(UserUtilTestSuite))
}

func (t *UserUtilTestSuite) TestUser() {
	userID := "userID"
	platform := "local"
	objectID := primitive.NewObjectID()
	userRet, err := user.CreateUser(userID, platform, objectID)
	t.Nil(err)
	t.NotNil(userRet)
	t.Equal(userRet.GmLevel, 1)

	userRet1, err1 := user.GetUser(userID)
	t.Nil(err1)
	t.NotNil(userRet1)
	t.Equal(userRet.UserID, userRet1.UserID)

	userRet, err = user.CreateUser(userID, platform, objectID)
	t.NotNil(err)
	t.Nil(userRet)

	err = user.DelUser(userID)
	t.Nil(err)

	userRet1, err1 = user.GetUser(userID)
	t.NotNil(err1)
	t.Nil(userRet1)

}

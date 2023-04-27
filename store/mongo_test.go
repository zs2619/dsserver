package store_test

import (
	"testing"

	"dsservices/test"

	"github.com/stretchr/testify/suite"
)

type MongoUtilTestSuite struct {
	test.BaseTestSuite
}

func TestMongoUtilTest(t *testing.T) {
	suite.Run(t, new(MongoUtilTestSuite))
}

func (t *MongoUtilTestSuite) TestMongo() {
}

package store_test

import (
	"testing"

	"dsservices/test"

	"github.com/stretchr/testify/suite"
)

type RedisUtilTestSuite struct {
	test.BaseTestSuite
}

func TestRedisUtilTest(t *testing.T) {
	suite.Run(t, new(RedisUtilTestSuite))
}

func (t *RedisUtilTestSuite) TestRedis() {
}

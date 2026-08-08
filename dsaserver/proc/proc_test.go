package proc_test

import (
	"dsservices/test"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ProcessUtilTestSuite struct {
	test.BaseTestSuite
}

func TestProcessUtilTest(t *testing.T) {
	suite.Run(t, new(ProcessUtilTestSuite))
}

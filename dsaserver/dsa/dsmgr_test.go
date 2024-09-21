package dsa

import (
	"dsservices/test"
	"testing"

	"github.com/stretchr/testify/suite"
)

type DSATestSuite struct {
	test.BaseTestSuite
}

func TestDSATest(t *testing.T) {
	suite.Run(t, new(DSATestSuite))
}

func (t *DSATestSuite) TestDSA() {
}

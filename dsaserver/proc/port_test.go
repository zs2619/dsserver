package proc_test

import (
	"dsservices/dsaserver/proc"
	"dsservices/test"
	"testing"

	"github.com/stretchr/testify/suite"
)

type PortUtilTestSuite struct {
	test.BaseTestSuite
}

func TestPortUtilTest(t *testing.T) {
	suite.Run(t, new(PortUtilTestSuite))
}

func (t *PortUtilTestSuite) TestPortMgr() {
	port0 := proc.GPortMgr.GetValidPort()
	t.Equal(port0, 30000)

	port1 := proc.GPortMgr.GetValidPort()
	t.Equal(port1, 30001)

	proc.GPortMgr.ReleasePort(port1)

	port2 := proc.GPortMgr.GetValidPort()
	t.Equal(port2, 30001)
}

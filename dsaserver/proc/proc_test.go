package proc_test

import (
	"dsservices/dsaserver/proc"
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

func (t *ProcessUtilTestSuite) TestProcess() {
	cmd := "shuai"
	dsID := "dsID"
	procinfo, err := proc.StartProc(dsID, cmd)
	t.Nil(err)
	t.NotNil(procinfo)
	t.Equal(procinfo.RealmCfgID, cmd)
	t.Equal(procinfo.Port, 30000)
	t.Equal(procinfo.Ip, "127.0.0.1")
}

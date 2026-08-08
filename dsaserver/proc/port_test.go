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
	// 用独立范围初始化，避免受其他测试顺序影响。
	mgr := proc.PortMgr{}
	mgr.Init(31000, 31100)

	port0 := mgr.GetValidPort()
	t.Equal(31000, port0)

	port1 := mgr.GetValidPort()
	t.Equal(31001, port1)

	mgr.ReleasePort(port1)

	port2 := mgr.GetValidPort()
	t.Equal(31001, port2)

	// MarkUsed 校正：推荐口与实际口不同时，推荐口应释放、实际口占用。
	mgr.ReleasePort(port0)
	mgr.ReleasePort(port2)
	ok := mgr.MarkUsed(31050)
	t.True(ok)
	// 31050 已占用，GetValidPort 不应再返回 31050
	next := mgr.GetValidPort()
	t.NotEqual(31050, next)
}

func (t *PortUtilTestSuite) TestMarkUsedConflict() {
	mgr := proc.PortMgr{}
	mgr.Init(32000, 32100)
	p := mgr.GetValidPort()
	t.Equal(32000, p)
	// 已在用，MarkUsed 返回 false
	ok := mgr.MarkUsed(32000)
	t.False(ok)
}

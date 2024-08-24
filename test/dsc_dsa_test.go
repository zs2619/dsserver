package test

import (
	"dsservices/dsaserver/dsa"
	"dsservices/dscserver/dsc"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type DscDsaTestSuite struct {
	BaseTestSuite
}

func TestProcessUtilTest(t *testing.T) {
	suite.Run(t, new(DscDsaTestSuite))
}

func (t *DscDsaTestSuite) TestDscDsa() {
	port := 30000
	agentID0 := "agentID0"
	agentID1 := "agentID1"
	grpcURL := "127.0.0.1:30000"

	dscServer, err := dsc.NewDSCServer(port)
	t.Nil(err)
	go dscServer.Run()

	agent0, err := dsa.NewDSAClient(agentID0, grpcURL)
	t.Nil(err)
	go agent0.RunStreamService()

	agent1, err := dsa.NewDSAClient(agentID1, grpcURL)
	go agent1.RunStreamService()
	t.Nil(err)
	time.Sleep(time.Second * 10)
}

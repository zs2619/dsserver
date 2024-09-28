package dsa

import (
	"bytes"
	"dsservices/kissnet"
	"dsservices/test"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type DSATestSuite struct {
	test.BaseTestSuite
}

func TestDSATest(t *testing.T) {
	suite.Run(t, new(DSATestSuite))
}

type dummyIConnection struct {
	kissnet.IConnection
}

// Close implements kissnet.IConnection.
func (d *dummyIConnection) Close() {
	panic("unimplemented")
}

// SendMsg implements kissnet.IConnection.
func (d *dummyIConnection) SendMsg(msg *bytes.Buffer) error {
	panic("unimplemented")
}

// getID implements kissnet.IConnection.
func (d *dummyIConnection) getID() int64 {
	panic("unimplemented")
}

// setID implements kissnet.IConnection.
func (d *dummyIConnection) setID(id int64) {
	panic("unimplemented")
}

// start implements kissnet.IConnection.
func (d *dummyIConnection) start() {
	panic("unimplemented")
}

func (t *DSATestSuite) TestDSA() {
	dsID := "dsID"
	realmCfgID := "realmCfgID"

	dsInfo := &DSInfo{
		DSConn:         nil,
		DsProcInfo:     nil,
		LastUpdateTime: time.Now(),
		CurrTime:       time.Now(),
		DSID:           dsID,
		DSState:        DS_CreatIng,
		RealmCfgID:     realmCfgID,
	}
	err := GDSInfoMgr.AddDS(dsInfo)
	t.Nil(err)

	retDsInfo := GDSInfoMgr.GetDSByID(dsID)
	t.NotNil(retDsInfo)

	err = retDsInfo.SetConnection(nil)
	t.NotNil(err)
	dummyconn := &dummyIConnection{}
	err = retDsInfo.SetConnection(dummyconn)
	t.Nil(err)
}

package dsa

import (
	"bytes"
	"dsservices/kissnet"
	"dsservices/test"
	"net"
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

type dummyConnection struct {
	kissnet.Connection
}

// Close implements kissnet.Connection.
func (d *dummyConnection) Close() {
	panic("unimplemented")
}

// SendMsg implements kissnet.Connection.
func (d *dummyConnection) SendMsg(msg *bytes.Buffer) error {
	panic("unimplemented")
}

// getID implements kissnet.Connection.
func (d *dummyConnection) getID() int64 {
	panic("unimplemented")
}

// setID implements kissnet.Connection.
func (d *dummyConnection) setID(id int64) {
	panic("unimplemented")
}

// start implements kissnet.Connection.
func (d *dummyConnection) start() {
	panic("unimplemented")
}

// RemoteAddr implements kissnet.Connection.
func (d *dummyConnection) RemoteAddr() net.Addr {
	return nil
}

func (t *DSATestSuite) TestDSA() {
	dsID := "dsID"
	realmCfgID := "realmCfgID"

	dsInfo := &DSInfo{
		DSConn:         nil,
		Port:           0,
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
	dummyconn := &dummyConnection{}
	err = retDsInfo.SetConnection(dummyconn)
	t.Nil(err)
}

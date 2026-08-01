package test

import (
	"context"
	"dsservices/config"
	"dsservices/dscserver/gamedsc"
	"dsservices/pb"
	"dsservices/store"
	"dsservices/store/dsastore"
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

// TestRealmAllocateIdempotent: no agent → async fail; ready → already_ready; concurrent → already_creating.
func (t *DscDsaTestSuite) TestRealmAllocateIdempotent() {
	ctx := context.Background()
	realmID := "realm-test-1"
	err := dsastore.WriteRealmSkeleton(ctx, realmID, "team1", "cfg1", []string{"u1"})
	t.Nil(err)

	srv := &gamedsc.RPCGameDscServer{}
	resp, err := srv.CreateRealm(ctx, &pb.RpcCreateRealmReq{
		RealmID:    realmID,
		TeamID:     "team1",
		RealmCfgID: "cfg1",
	})
	t.Nil(err)
	t.Equal(pb.CreateRealmStatus_CREATE_REALM_ACCEPTED, resp.Status)

	// concurrent claim while allocating
	respDup, err := srv.CreateRealm(ctx, &pb.RpcCreateRealmReq{
		RealmID:    realmID,
		TeamID:     "team1",
		RealmCfgID: "cfg1",
	})
	t.Nil(err)
	t.Equal(pb.CreateRealmStatus_CREATE_REALM_ALREADY_CREATING, respDup.Status)

	// no agent → async fail
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		info, _ := dsastore.GetRealm(ctx, realmID)
		if info != nil && info.State == dsastore.RealmStateFail {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	info, err := dsastore.GetRealm(ctx, realmID)
	t.Nil(err)
	t.NotNil(info)
	t.Equal(dsastore.RealmStateFail, info.State)

	_ = dsastore.MarkRealmReady(ctx, realmID, "ds1", "127.0.0.1:30000")
	resp2, err := srv.CreateRealm(ctx, &pb.RpcCreateRealmReq{
		RealmID:    realmID,
		TeamID:     "team1",
		RealmCfgID: "cfg1",
	})
	t.Nil(err)
	t.Equal(pb.CreateRealmStatus_CREATE_REALM_ALREADY_READY, resp2.Status)
	t.Equal("127.0.0.1:30000", resp2.DsAddr)

	_ = config.GameConfig
	_ = store.RedisClient
}

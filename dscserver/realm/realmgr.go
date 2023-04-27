package realm

import (
	"fmt"
	"sync"
)

var RealmInfoMap sync.Map

type RealmInfo struct {
	RealmID    string
	RealmCfgID string
	Addr       string
	TeamID     []string
}

func GetRealmInfo(realmCfgID string) (retRealmInfo *RealmInfo, err error) {
	//TODO:Team分配给ds的策略
	return nil, fmt.Errorf("realm nil")
}

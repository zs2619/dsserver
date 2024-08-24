package realm

import (
	"sync"
)

var DSInfoMap sync.Map

type DSInfo struct {
	DSID       string
	RealmCfgID string
	Addr       string
}

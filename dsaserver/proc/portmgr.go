package proc

import "sync"

const poolSize = 30000

// 分配范围30000 -60000
type PortMgr struct {
	PortPool    [poolSize]int
	portMgrMute sync.Mutex
}

var GPortMgr PortMgr

func (this *PortMgr) GetValidPort() int {
	this.portMgrMute.Lock()
	defer this.portMgrMute.Unlock()

	for k, v := range this.PortPool {
		if v == 0 {
			this.PortPool[k] = 1
			return poolSize + k
		}
	}

	return 0
}

func (this *PortMgr) ReleasePort(port int) {
	this.portMgrMute.Lock()
	defer this.portMgrMute.Unlock()
	index := port - poolSize
	this.PortPool[index] = 0
}

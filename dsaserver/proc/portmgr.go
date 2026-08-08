package proc

import "sync"

// PortMgr 在 [min, max] 范围内分配端口。
// container 模式下范围由 env GAME_PORT_MIN/MAX 注入（本 Deployment 段），
// process 模式下用 config 全段（默认 30000-60000）。
type PortMgr struct {
	min          int
	max          int
	portInUse    map[int]bool
	portMgrMute  sync.Mutex
	initialized  bool
}

var GPortMgr PortMgr

// Init 设定分配范围；只应调用一次。未调用时用默认 30000-60000。
func (p *PortMgr) Init(min, max int) {
	p.portMgrMute.Lock()
	defer p.portMgrMute.Unlock()
	if p.initialized {
		return
	}
	if min <= 0 {
		min = 30000
	}
	if max <= 0 || max < min {
		max = 60000
	}
	p.min = min
	p.max = max
	if p.portInUse == nil {
		p.portInUse = make(map[int]bool)
	}
	p.initialized = true
}

func (p *PortMgr) ensureInit() {
	if !p.initialized {
		p.Init(30000, 60000)
	}
}

// GetValidPort 返回一个空闲口并标记为已用；无可用返回 0。
func (p *PortMgr) GetValidPort() int {
	p.ensureInit()
	p.portMgrMute.Lock()
	defer p.portMgrMute.Unlock()
	for port := p.min; port <= p.max; port++ {
		if !p.portInUse[port] {
			p.portInUse[port] = true
			return port
		}
	}
	return 0
}

// MarkUsed 标记某口为已用（DS 内换口后，CreateOK 报最终口，DSA 据此校正记账）。
// 若该口已在用返回 false。
func (p *PortMgr) MarkUsed(port int) bool {
	p.ensureInit()
	p.portMgrMute.Lock()
	defer p.portMgrMute.Unlock()
	if port < p.min || port > p.max {
		return false
	}
	if p.portInUse[port] {
		return false
	}
	p.portInUse[port] = true
	return true
}

func (p *PortMgr) ReleasePort(port int) {
	if port <= 0 {
		return
	}
	p.ensureInit()
	p.portMgrMute.Lock()
	defer p.portMgrMute.Unlock()
	delete(p.portInUse, port)
}

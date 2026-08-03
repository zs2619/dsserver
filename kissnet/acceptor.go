package kissnet

import (
	"fmt"
	"time"
)

type Acceptor interface {
	Start() error
	Run() error
	Close() error
	WaitConnectionsClosed(timeout time.Duration) bool
	IsRunning() bool
}

type acceptorBase struct {
	connMgr ConnectionMgr
	SessionCallback
}

// NewAcceptor 按类型创建监听器，目前仅支持 tcp。
func NewAcceptor(acceptorType string, port int, cb SessionCallback, ne *NetEvent) (Acceptor, error) {
	if acceptorType == "ws" {
		return nil, fmt.Errorf("ws acceptor not implemented")
	} else if acceptorType == "tcp" {
		return NewTCPAcceptor(port, cb, ne)
	}
	return nil, fmt.Errorf("config AcceptorType error")
}

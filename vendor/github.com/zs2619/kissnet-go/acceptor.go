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

// NewAcceptor 按类型创建监听器，支持 tcp / ws / websocket。
// ws 使用 DefaultWSPath（默认 "/ws"），BinaryMessage 一条业务包（无 TCP 长度头）。
func NewAcceptor(acceptorType string, port int, cb SessionCallback, ne *NetEvent) (Acceptor, error) {
	switch acceptorType {
	case "tcp":
		return NewTCPAcceptor(port, cb, ne)
	case "ws", "websocket":
		return NewWSAcceptor(port, cb, ne)
	default:
		return nil, fmt.Errorf("config AcceptorType error")
	}
}

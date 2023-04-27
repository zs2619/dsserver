package kissnet

import (
	"fmt"
)

type IAcceptor interface {
	Run() error
	Close() error
	IsRunning() bool
}

type Acceptor struct {
	CMgr ConnectionMgr
	CallBack
}

func AcceptorFactory(acceptorType string, port int, cb ConnectionCB) (IAcceptor, error) {
	if acceptorType == "ws" {
		return nil, nil
	} else if acceptorType == "tcp" {
		return NewTcpAcceptor(port, cb)
	} else {
		return nil, fmt.Errorf("config AcceptorType error")
	}
}

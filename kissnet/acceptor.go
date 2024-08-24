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
	SessionCallBack
}

func AcceptorFactory(acceptorType string, port int, cb SessionCallBack) (IAcceptor, error) {
	if acceptorType == "ws" {
		// return NewWSAcceptor(port, cb)
		return nil, nil
	} else if acceptorType == "tcp" {
		return NewTcpAcceptor(port, cb)
	} else {
		return nil, fmt.Errorf("config AcceptorType error")
	}
}

package kissnet

import (
	"fmt"
	"net"
	"sync/atomic"

	"github.com/sirupsen/logrus"
)

type TcpAcceptor struct {
	Acceptor
	listener *net.TCPListener
	Addr     *net.TCPAddr
	running  int32
}

func NewTcpAcceptor(port int, cb SessionCallBack) (*TcpAcceptor, error) {
	if cb == nil {
		return nil, fmt.Errorf("cb nil")
	}
	ep := fmt.Sprintf("0.0.0.0:%d", port)
	tcpAddr, err := net.ResolveTCPAddr("tcp", ep)
	if err != nil {
		return nil, err
	}
	ln, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return nil, err
	}
	return &TcpAcceptor{Addr: tcpAddr, listener: ln, Acceptor: Acceptor{SessionCallBack: cb}}, nil
}

func (acceptor *TcpAcceptor) Run() error {
	acceptor.accept()
	return nil
}

func (acceptor *TcpAcceptor) accept() {
	atomic.StoreInt32(&acceptor.running, 1)
	for {
		tcpConn, err := acceptor.listener.AcceptTCP()
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{}).Error("TcpAcceptor::accept error")
			return
		}
		conn := NewConnection(tcpConn, acceptor.Acceptor.SessionCallBack)
		go conn.start()
	}
}

func (acceptor *TcpAcceptor) Close() error {
	acceptor.listener.Close()
	atomic.StoreInt32(&acceptor.running, 0)
	return nil
}

func (acceptor *TcpAcceptor) IsRunning() bool {
	return atomic.LoadInt32(&acceptor.running) > 0
}

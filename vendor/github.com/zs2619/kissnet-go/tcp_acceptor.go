package kissnet

import (
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

type TCPAcceptor struct {
	acceptorBase
	listener *net.TCPListener
	Addr     *net.TCPAddr
	running  int32
	ne       *NetEvent
}

func NewTCPAcceptor(port int, cb SessionCallback, ne *NetEvent) (*TCPAcceptor, error) {
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
	return &TCPAcceptor{Addr: tcpAddr, listener: ln, ne: ne, acceptorBase: acceptorBase{SessionCallback: cb}}, nil
}

// Start 非阻塞启动 accept；出错时关闭连接与 EventLoop。
func (a *TCPAcceptor) Start() error {
	go func() {
		if err := a.accept(); err != nil {
			logrus.WithError(err).Error("TCPAcceptor accept exited, shutting down")
			_ = a.Close()
			if !a.WaitConnectionsClosed(10 * time.Second) {
				logrus.Warn("not all connections closed within grace period (accept error path)")
			}
			if a.ne != nil {
				a.ne.Close()
			}
		}
	}()
	return nil
}

// Run 阻塞 accept，出错时返回 error。
func (a *TCPAcceptor) Run() error {
	return a.accept()
}

func (a *TCPAcceptor) accept() error {
	atomic.StoreInt32(&a.running, 1)
	for {
		tcpConn, err := a.listener.AcceptTCP()
		if err != nil {
			if atomic.LoadInt32(&a.running) == 0 {
				return nil
			}
			logrus.WithError(err).Error("TCPAcceptor::accept error")
			return err
		}
		conn := NewConnection(tcpConn, a.SessionCallback, a.ne)
		if !a.connMgr.add(conn) {
			_ = tcpConn.Close()
			continue
		}
		go func(c Connection) {
			c.start()
			a.connMgr.del(c)
		}(conn)
	}
}

func (a *TCPAcceptor) Close() error {
	atomic.StoreInt32(&a.running, 0)
	_ = a.listener.Close()
	a.connMgr.CloseAll()
	return nil
}

// WaitConnectionsClosed 等待连接退出（含 OnDisconnectCB 入队），超时返回 false。
func (a *TCPAcceptor) WaitConnectionsClosed(timeout time.Duration) bool {
	return a.connMgr.WaitEmpty(timeout)
}

func (a *TCPAcceptor) IsRunning() bool {
	return atomic.LoadInt32(&a.running) > 0
}

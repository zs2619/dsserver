package kissnet

import (
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// DefaultWSPath is the HTTP upgrade path used by NewAcceptor("ws", ...).
var DefaultWSPath = "/ws"

type WSAcceptor struct {
	acceptorBase
	addr    string
	path    string
	server  *http.Server
	running int32
	ne      *NetEvent
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  8192,
	WriteBufferSize: 8192,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// NewWSAcceptor creates a WebSocket listener on port with DefaultWSPath.
func NewWSAcceptor(port int, cb SessionCallback, ne *NetEvent) (*WSAcceptor, error) {
	return NewWSAcceptorPath(port, DefaultWSPath, cb, ne)
}

// NewWSAcceptorPath creates a WebSocket listener; path is the upgrade URL path (e.g. "/ws").
func NewWSAcceptorPath(port int, path string, cb SessionCallback, ne *NetEvent) (*WSAcceptor, error) {
	if cb == nil {
		return nil, fmt.Errorf("cb nil")
	}
	path = normalizeWSPath(path)
	addrStr := fmt.Sprintf(":%d", port)
	return &WSAcceptor{
		addr:         addrStr,
		path:         path,
		server:       &http.Server{Addr: addrStr},
		ne:           ne,
		acceptorBase: acceptorBase{SessionCallback: cb},
	}, nil
}

func normalizeWSPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return "/ws"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return path
}

func (a *WSAcceptor) Path() string { return a.path }

func (a *WSAcceptor) Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc(a.path, a.handleUpgrade)
	a.server.Handler = mux
	atomic.StoreInt32(&a.running, 1)
	err := a.server.ListenAndServe()
	atomic.StoreInt32(&a.running, 0)
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func (a *WSAcceptor) handleUpgrade(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.WithError(err).Error("WSAcceptor upgrade error")
		return
	}
	ws := newWSConnection(c, a.SessionCallback, a.ne)
	if !a.connMgr.add(ws) {
		_ = c.Close()
		return
	}
	go func(conn Connection) {
		conn.start()
		a.connMgr.del(conn)
	}(ws)
}

// Start 非阻塞启动；出错时关闭连接与 EventLoop。
func (a *WSAcceptor) Start() error {
	go func() {
		if err := a.Run(); err != nil {
			logrus.WithError(err).Error("WSAcceptor Run exited, shutting down")
			_ = a.Close()
			if !a.WaitConnectionsClosed(10 * time.Second) {
				logrus.Warn("not all ws connections closed within grace period")
			}
			if a.ne != nil {
				a.ne.Close()
			}
		}
	}()
	return nil
}

func (a *WSAcceptor) Close() error {
	atomic.StoreInt32(&a.running, 0)
	if a.server != nil {
		_ = a.server.Close()
	}
	a.connMgr.CloseAll()
	return nil
}

func (a *WSAcceptor) WaitConnectionsClosed(timeout time.Duration) bool {
	return a.connMgr.WaitEmpty(timeout)
}

func (a *WSAcceptor) IsRunning() bool {
	return atomic.LoadInt32(&a.running) > 0
}

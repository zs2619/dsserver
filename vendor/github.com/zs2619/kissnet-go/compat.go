package kissnet

import "sync"

// Compatibility aliases / helpers for older kissnet-go callers.

// IConnection is an alias of Connection (historical name).
type IConnection = Connection

// IAcceptor is an alias of Acceptor (historical name).
type IAcceptor = Acceptor

// ConnectionCB is the legacy single-callback style.
// msg == nil means disconnect.
type ConnectionCB func(Connection, []byte) error

type funcSessionCB struct {
	f ConnectionCB
}

func (c funcSessionCB) OnConnectCB(Connection) error { return nil }

func (c funcSessionCB) OnDisconnectCB(conn Connection) error {
	if c.f == nil {
		return nil
	}
	return c.f(conn, nil)
}

func (c funcSessionCB) OnMsgCB(conn Connection, msg []byte) error {
	if c.f == nil {
		return nil
	}
	return c.f(conn, msg)
}

// AdaptConnectionCB wraps a legacy ConnectionCB as SessionCallback.
func AdaptConnectionCB(cb ConnectionCB) SessionCallback {
	return funcSessionCB{f: cb}
}

// G_NetEvent is the legacy package-level NetEvent used by AcceptorFactory / TcpConnector.
// Prefer creating a NetEvent and passing it explicitly (dsserver style).
var (
	G_NetEvent  *NetEvent
	gNetEventMu sync.Mutex
)

func ensureGNetEvent() *NetEvent {
	gNetEventMu.Lock()
	defer gNetEventMu.Unlock()
	if G_NetEvent == nil {
		G_NetEvent = NewNetEvent()
	}
	return G_NetEvent
}

// SetGNetEvent installs the NetEvent used by legacy AcceptorFactory / TcpConnector.
func SetGNetEvent(ne *NetEvent) {
	gNetEventMu.Lock()
	G_NetEvent = ne
	gNetEventMu.Unlock()
}

// AcceptorFactory is the legacy factory (no NetEvent argument).
func AcceptorFactory(acceptorType string, port int, cb ConnectionCB) (Acceptor, error) {
	return NewAcceptor(acceptorType, port, AdaptConnectionCB(cb), ensureGNetEvent())
}

// TcpConnector is the legacy dial helper name.
func TcpConnector(addr string, cb ConnectionCB) (Connection, error) {
	return TCPConnector(addr, AdaptConnectionCB(cb), ensureGNetEvent())
}

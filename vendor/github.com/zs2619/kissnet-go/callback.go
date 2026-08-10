package kissnet

type SessionOnConnectCallback interface {
	OnConnectCB(Connection) error
}
type SessionOnDisconnectCallback interface {
	OnDisconnectCB(Connection) error
}
type SessionCallback interface {
	SessionOnConnectCallback
	SessionOnDisconnectCallback
	OnMsgCB(Connection, []byte) error
}

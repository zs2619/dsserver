package kissnet

type SessionOnConectCallBack interface {
	OnConectCB(IConnection) error
}
type SessionOnDisConectCallBack interface {
	OnDisConectCB(IConnection) error
}
type SessionCallBack interface {
	SessionOnConectCallBack
	SessionOnDisConectCallBack
	OnMsgCB(IConnection, []byte) error
}

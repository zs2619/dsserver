package kissnet

type ConnectionCB func(IConnection, []byte) error
type CallBack struct {
	ConnectionCB ConnectionCB
}

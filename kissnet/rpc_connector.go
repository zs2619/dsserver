package kissnet

import (
	"bytes"
	"net"
)

type RpcConnector struct {
	serviceName string
	conn        IConnection
	addr        string
}

func (this *RpcConnector) start() error {
	this.connect()
	return nil
}

func (this *RpcConnector) connect() error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", this.addr)
	if err != nil {
		return err
	}

	c, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return err
	}
	this.conn = NewConnection(c, nil)
	this.conn.start()
	return nil
}
func (this *RpcConnector) sendMsg(buf *bytes.Buffer) error {
	return nil
}

func (this *RpcConnector) Close() error {
	return nil
}

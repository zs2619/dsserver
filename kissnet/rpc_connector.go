package kissnet

import (
	"bytes"
	"net"
)

type RPCConnector struct {
	serviceName string
	conn        Connection
	addr        string
	ne          *NetEvent
}

func (c *RPCConnector) start() error {
	return c.connect()
}

func (c *RPCConnector) connect() error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", c.addr)
	if err != nil {
		return err
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return err
	}
	c.conn = NewConnection(conn, nil, c.ne)
	c.conn.start()
	return nil
}

func (c *RPCConnector) sendMsg(buf *bytes.Buffer) error {
	if c.conn == nil {
		return errConnClosed
	}
	return c.conn.SendMsg(buf)
}

func (c *RPCConnector) Close() error {
	if c.conn == nil {
		return nil
	}
	c.conn.Close()
	return nil
}

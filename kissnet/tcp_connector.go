package kissnet

import (
	"fmt"
	"net"
)

func TcpConnector(addr string, cb SessionCallBack) (IConnection, error) {
	if cb == nil {
		return nil, fmt.Errorf("cb nil")
	}
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, err
	}
	c := NewConnection(conn, cb)
	go c.start()
	return c, nil
}

package kissnet

import (
	"fmt"
	"net"
)

func TCPConnector(addr string, cb SessionCallback, ne *NetEvent) (Connection, error) {
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
	c := NewConnection(conn, cb, ne)
	go c.start()
	return c, nil
}

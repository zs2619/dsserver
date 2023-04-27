package kissnet

import "net"

func TcpConnector(addr string, cb ConnectionCB) (IConnection, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, err
	}
	c := NewConnection(conn, &CallBack{ConnectionCB: cb})
	return c, nil
}

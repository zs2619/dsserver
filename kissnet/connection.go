package kissnet

import (
	"bytes"
	"net"
)

// Connection 网络连接。
type Connection interface {
	start()
	getID() int64
	setID(id int64)
	SendMsg(msg *bytes.Buffer) error
	Close()
	RemoteAddr() net.Addr
}

// PeerIP 返回对端 IP（不含端口）；不可用时返回空串。
func PeerIP(c Connection) string {
	if c == nil {
		return ""
	}
	addr := c.RemoteAddr()
	if addr == nil {
		return ""
	}
	host, _, err := net.SplitHostPort(addr.String())
	if err != nil {
		return addr.String()
	}
	return host
}

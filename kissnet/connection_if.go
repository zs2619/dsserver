package kissnet

import "bytes"

type IConnection interface {
	start()
	getID() int64
	setID(id int64)
	SendMsg(msg *bytes.Buffer) error
	Close()
}

package kissnet

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"sync"
	"sync/atomic"

	"github.com/sirupsen/logrus"
)

const MsgHeaderMaxSize = 2

type Connection struct {
	id           int64
	conn         *net.TCPConn
	exitSync     sync.WaitGroup
	cb           *CallBack
	isClose      int32
	sendCh       chan *bytes.Buffer
	lastPingTime int64
}

func NewConnection(conn *net.TCPConn, cb *CallBack) IConnection {
	c := &Connection{
		conn:    conn,
		isClose: 0,
		cb:      cb,
		sendCh:  make(chan *bytes.Buffer, 1024),
	}
	return c
}

func (this *Connection) getID() int64 {
	return this.id
}
func (this *Connection) setID(id int64) {
	this.id = id
}

func (this *Connection) IsClose() bool { return atomic.LoadInt32(&this.isClose) > 0 }

func (this *Connection) Close() {
	this.cb.ConnectionCB(this, nil)
	this.SendMsg(nil)

	this.exitSync.Wait()
	atomic.StoreInt32(&this.isClose, 1)

	close(this.sendCh)
	if this.conn != nil {
		this.conn.Close()
		this.conn = nil
	}
	logrus.WithFields(logrus.Fields{"id:": this.id}).Info("TcpConnection  Close")
}

func (this *Connection) start() {
	this.conn.SetNoDelay(true)
	this.conn.SetKeepAlive(true)

	//同步退出 goroutine
	this.exitSync.Add(2)

	//开启读写 goroutine
	go this.recvMsgLoop()
	go this.sendMsgLoop()
}

func (this *Connection) SendMsg(msg *bytes.Buffer) error {
	if this.IsClose() {
		//关闭不能发送消息
		return nil
	}
	//推入发送循环
	this.sendCh <- msg
	return nil
}

func (this *Connection) sendMsgLoop() {
	for msg := range this.sendCh {
		if msg == nil || this.conn == nil {
			break
		}

		msgLen := uint16(msg.Len())
		buf := make([]byte, MsgHeaderMaxSize)
		binary.LittleEndian.PutUint16(buf, msgLen)
		_, err := this.conn.Write(buf)
		if err != nil {
			break
		}
		_, err = this.conn.Write(msg.Bytes())
		if err != nil {
			break
		}
	}
	//关闭socket 从读操作退出
	this.exitSync.Done()
}

func (this *Connection) recvMsgLoop() {
	var err error
	var msgLen int
	msgHeader := make([]byte, MsgHeaderMaxSize)
	for {
		_, err = io.ReadFull(this.conn, msgHeader)
		if err != nil {
			break
		}
		msgLen = int(binary.LittleEndian.Uint16(msgHeader))
		if msgLen <= 0 || msgLen > 65535 {
			break
		}

		msgBody := make([]byte, msgLen)
		_, err = io.ReadFull(this.conn, msgBody)
		if err != nil {
			break
		}
		this.cb.ConnectionCB(this, msgBody)
	}
	this.exitSync.Done()
	// 退出处理
	this.Close()
}

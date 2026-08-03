package kissnet

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const msgHeaderMaxSize = 2

var errConnClosed = errors.New("connection closed")

// tcpConnection TCP 连接。
// recv/send 只做 IO；OnConnect/OnMsg/OnDisconnect 投递到 NetEvent 串行执行。
//
// 关闭顺序：
//  1. Close 只关读侧，唤醒 recv
//  2. recv 退出后 close(sendCh)，send 排空后退出
//  3. send 退出时完整关闭 socket
//  4. start 在收发结束后投递 OnDisconnectCB
type tcpConnection struct {
	id           int64
	conn         *net.TCPConn
	exitSync     sync.WaitGroup
	cb           SessionCallback
	ne           *NetEvent // nil 时回调在当前协程直接执行
	closing      int32
	sendCh       chan *bytes.Buffer
	sendMu       sync.Mutex
	sendChClosed bool
}

// NewConnection 创建连接，返回 Connection 接口。
func NewConnection(conn *net.TCPConn, cb SessionCallback, ne *NetEvent) Connection {
	return &tcpConnection{
		conn:   conn,
		cb:     cb,
		ne:     ne,
		sendCh: make(chan *bytes.Buffer, 1024),
	}
}

func (c *tcpConnection) RemoteAddr() net.Addr {
	if c.conn == nil {
		return nil
	}
	return c.conn.RemoteAddr()
}

func (c *tcpConnection) getID() int64 {
	return c.id
}
func (c *tcpConnection) setID(id int64) {
	c.id = id
}

func (c *tcpConnection) isClose() bool {
	return atomic.LoadInt32(&c.closing) != 0
}

func (c *tcpConnection) dispatch(cb func()) {
	if c.ne != nil {
		c.ne.ScheduleMsg(cb)
	} else if cb != nil {
		cb()
	}
}

func (c *tcpConnection) dispatchTimeout(cb func(), timeout time.Duration) {
	if c.ne != nil {
		c.ne.ScheduleMsgTimeout(cb, timeout)
	} else if cb != nil {
		cb()
	}
}

// Close 关闭读侧以唤醒 recv；幂等。完整关闭由 send 退出时完成。
func (c *tcpConnection) Close() {
	if atomic.SwapInt32(&c.closing, 1) != 0 {
		return
	}
	if c.conn != nil {
		_ = c.conn.CloseRead()
		_ = c.conn.SetReadDeadline(time.Now())
	}
	logrus.WithFields(logrus.Fields{"id": c.id}).Info("tcpConnection Close (close read)")
}

// notifySendExit 关闭 sendCh，通知发送循环退出。
func (c *tcpConnection) notifySendExit() {
	atomic.StoreInt32(&c.closing, 1)
	c.sendMu.Lock()
	defer c.sendMu.Unlock()
	if c.sendChClosed {
		return
	}
	c.sendChClosed = true
	close(c.sendCh)
}

func (c *tcpConnection) start() {
	defer func() {
		if err := recover(); err != nil {
			logrus.WithFields(logrus.Fields{"error": err}).Error("connection start error")
		}
	}()

	// Close 可能先于 start：此时不启收发、不投回调，自行完整关闭 socket。
	if c.isClose() {
		if c.conn != nil {
			_ = c.conn.Close()
		}
		return
	}

	c.conn.SetNoDelay(true)
	c.conn.SetKeepAlive(true)

	c.dispatchTimeout(func() {
		if cb, ok := c.cb.(SessionOnConnectCallback); ok {
			_ = cb.OnConnectCB(c)
		}
	}, 100*time.Millisecond)

	c.exitSync.Add(2)
	go c.recvMsgLoop()
	go c.sendMsgLoop()
	c.exitSync.Wait()

	c.dispatchTimeout(func() {
		if cb, ok := c.cb.(SessionOnDisconnectCallback); ok {
			_ = cb.OnDisconnectCB(c)
		}
	}, time.Second)
}

func (c *tcpConnection) SendMsg(msg *bytes.Buffer) error {
	if msg == nil {
		return nil
	}
	if msg.Len() > 0xFFFF {
		return errors.New("message too large")
	}
	c.sendMu.Lock()
	defer c.sendMu.Unlock()
	if c.sendChClosed || c.isClose() {
		return errConnClosed
	}
	select {
	case c.sendCh <- msg:
		return nil
	default:
		return errors.New("send buffer full")
	}
}

func (c *tcpConnection) sendMsgLoop() {
	defer func() {
		if err := recover(); err != nil {
			if e, ok := err.(error); ok {
				logrus.WithError(errors.WithStack(e)).Error("sendMsgLoop error")
			}
		}
		if c.conn != nil {
			_ = c.conn.Close()
		}
		c.exitSync.Done()
	}()

	for msg := range c.sendCh {
		head := make([]byte, msgHeaderMaxSize)
		binary.LittleEndian.PutUint16(head, uint16(msg.Len()))
		if _, err := (&net.Buffers{head, msg.Bytes()}).WriteTo(c.conn); err != nil {
			logrus.WithError(err).Error("c.conn.Write packet")
			return
		}
	}
}

func (c *tcpConnection) recvMsgLoop() {
	defer func() {
		if err := recover(); err != nil {
			if e, ok := err.(error); ok {
				logrus.WithError(errors.WithStack(e)).Error("recvMsgLoop error")
			}
		}
		c.notifySendExit()
		c.exitSync.Done()
	}()

	msgHeader := make([]byte, msgHeaderMaxSize)
	for {
		_, err := io.ReadFull(c.conn, msgHeader)
		if err != nil {
			if err != io.EOF {
				logrus.WithError(err).Error("io.ReadFull head")
			}
			return
		}
		msgLen := int(binary.LittleEndian.Uint16(msgHeader))
		if msgLen == 0 {
			logrus.WithFields(logrus.Fields{"msgLen": msgLen}).Error("msg len error")
			return
		}

		msgBody := make([]byte, msgLen)
		_, err = io.ReadFull(c.conn, msgBody)
		if err != nil {
			logrus.WithError(err).Error("io.ReadFull body")
			return
		}
		body := msgBody
		c.dispatch(func() {
			if c.cb != nil {
				_ = c.cb.OnMsgCB(c, body)
			}
		})
	}
}

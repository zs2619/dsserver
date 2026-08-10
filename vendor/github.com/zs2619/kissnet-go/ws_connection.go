package kissnet

import (
	"bytes"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const recvWait = 120 * time.Second

// wsConnection WebSocket 连接。
// 每个 BinaryMessage 即一条业务消息（不含 TCP 的 uint16 长度头）。
type wsConnection struct {
	id           int64
	conn         *websocket.Conn
	cb           SessionCallback
	ne           *NetEvent
	closing      int32
	sendCh       chan *bytes.Buffer
	sendMu       sync.Mutex
	sendChClosed bool
	exitSync     sync.WaitGroup
}

func newWSConnection(conn *websocket.Conn, cb SessionCallback, ne *NetEvent) Connection {
	c := &wsConnection{
		conn:   conn,
		cb:     cb,
		ne:     ne,
		sendCh: make(chan *bytes.Buffer, 1024),
	}
	c.conn.SetPingHandler(func(message string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(recvWait))
		err := c.conn.WriteControl(websocket.PongMessage, []byte(message), time.Now().Add(time.Second))
		if err == websocket.ErrCloseSent {
			return nil
		}
		if e, ok := err.(net.Error); ok && e.Temporary() {
			return nil
		}
		return err
	})
	return c
}

func (c *wsConnection) getID() int64  { return c.id }
func (c *wsConnection) setID(id int64) { c.id = id }

func (c *wsConnection) RemoteAddr() net.Addr {
	if c.conn == nil {
		return nil
	}
	return c.conn.RemoteAddr()
}

func (c *wsConnection) isClose() bool {
	return atomic.LoadInt32(&c.closing) != 0
}

func (c *wsConnection) dispatch(cb func()) {
	if c.ne != nil {
		c.ne.ScheduleMsg(cb)
	} else if cb != nil {
		cb()
	}
}

func (c *wsConnection) dispatchTimeout(cb func(), timeout time.Duration) {
	if c.ne != nil {
		c.ne.ScheduleMsgTimeout(cb, timeout)
	} else if cb != nil {
		cb()
	}
}

func (c *wsConnection) Close() {
	if !atomic.CompareAndSwapInt32(&c.closing, 0, 1) {
		return
	}
	if c.conn != nil {
		_ = c.conn.Close()
	}
}

func (c *wsConnection) notifySendExit() {
	c.sendMu.Lock()
	defer c.sendMu.Unlock()
	if c.sendChClosed {
		return
	}
	c.sendChClosed = true
	close(c.sendCh)
}

func (c *wsConnection) SendMsg(msg *bytes.Buffer) error {
	if msg == nil {
		return nil
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

func (c *wsConnection) start() {
	defer func() {
		if err := recover(); err != nil {
			logrus.WithFields(logrus.Fields{"error": err}).Error("ws connection start error")
		}
	}()

	if c.isClose() {
		if c.conn != nil {
			_ = c.conn.Close()
		}
		return
	}

	c.dispatchTimeout(func() {
		if c.cb != nil {
			_ = c.cb.OnConnectCB(c)
		}
	}, 100*time.Millisecond)

	c.exitSync.Add(2)
	go c.recvLoop()
	go c.sendLoop()
	c.exitSync.Wait()

	c.dispatchTimeout(func() {
		if c.cb != nil {
			_ = c.cb.OnDisconnectCB(c)
		}
	}, time.Second)
}

func (c *wsConnection) sendLoop() {
	defer func() {
		if c.conn != nil {
			_ = c.conn.Close()
		}
		c.exitSync.Done()
	}()

	for msg := range c.sendCh {
		if c.conn == nil {
			return
		}
		if err := c.conn.WriteMessage(websocket.BinaryMessage, msg.Bytes()); err != nil {
			return
		}
	}
}

func (c *wsConnection) recvLoop() {
	defer func() {
		c.notifySendExit()
		c.exitSync.Done()
	}()

	for {
		if c.conn == nil {
			return
		}
		_ = c.conn.SetReadDeadline(time.Now().Add(recvWait))
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
		body := append([]byte(nil), data...)
		c.dispatch(func() {
			if c.cb != nil {
				_ = c.cb.OnMsgCB(c, body)
			}
		})
	}
}

package main

import (
	"bytes"
	"github.com/zs2619/kissnet-go"
	"dsservices/pb"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

// dstemplate 模拟 UE 模板 DS：
//  1. 连上 DSA 模板口，发 TemplateReady；
//  2. 收到 ForkInstance 后 fork 一个实例（此处用 goroutine 模拟）：
//     - 实例连上 DSA 实例口，发 CreateOK（dsAddr=advertiseIP:port）；
//     - 实例在游戏口上 listen，供玩家/联调连接；
//     - 收到 StopDS 或连接断开即退出。
//
// 环境变量：
//   DSA_TEMPLATE_ADDR  DSA 模板口地址（模板连此），如 127.0.0.1:11001
//   DSA_DS_ADDR        DSA 实例口地址（实例连此），如 127.0.0.1:11000
//   DS_ADVERTISE_IP    玩家可达 IP（默认 127.0.0.1）
func main() {
	logrus.Info("dstemplate start")

	templateAddr := os.Getenv("DSA_TEMPLATE_ADDR")
	if templateAddr == "" {
		templateAddr = "127.0.0.1:11001"
	}
	dsAddr := os.Getenv("DSA_DS_ADDR")
	if dsAddr == "" {
		dsAddr = "127.0.0.1:11000"
	}
	advertiseIP := os.Getenv("DS_ADVERTISE_IP")
	if advertiseIP == "" {
		advertiseIP = "127.0.0.1"
	}

	event := kissnet.NewNetEvent()
	cb := &templateCB{dsAddr: dsAddr, advertiseIP: advertiseIP, event: event, templateAddr: templateAddr}

	go event.EventLoop()

	// 首次连接带重试（DSA 可能还没起好）；之后断开由 OnDisconnectCB 触发重连。
	cb.connectLoop(true)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	logrus.Info("dstemplate shutdown")
	cb.close()
	event.Close()
}

// templateCB 处理模板口连接（收 ForkInstance）。
type templateCB struct {
	connMu       sync.Mutex
	conn         kissnet.Connection
	dsAddr       string // DSA 实例口地址
	advertiseIP  string
	event        *kissnet.NetEvent
	templateAddr string
	mu           sync.Mutex
	instances    map[string]*instanceCtx
	closed       int32 // atomic; 1 后不再重连
	closeOnce    sync.Once
}

func (c *templateCB) setConn(conn kissnet.Connection) {
	c.connMu.Lock()
	c.conn = conn
	c.connMu.Unlock()
}

func (c *templateCB) getConn() kissnet.Connection {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	return c.conn
}

// connectLoop 重试连接 DSA 模板口，成功后发 TemplateReady。
// first=true 时阻塞到首次成功；之后断开由 OnDisconnectCB 异步重连。
func (c *templateCB) connectLoop(first bool) {
	backoff := time.Second
	maxBackoff := 10 * time.Second
	for {
		if atomic.LoadInt32(&c.closed) == 1 {
			return
		}
		conn, err := kissnet.TCPConnector(c.templateAddr, c, c.event)
		if err == nil {
			c.setConn(conn)
			c.sendTemplateReady()
			logrus.Info("template connected to DSA, TemplateReady sent")
			return
		}
		if first {
			logrus.WithError(err).Warn("connect DSA template port failed, retrying")
		} else {
			logrus.WithError(err).Warn("reconnect DSA template port failed, retrying")
		}
		time.Sleep(backoff)
		if backoff < maxBackoff {
			backoff += time.Second
		}
	}
}

func (c *templateCB) ensure() {
	c.mu.Lock()
	if c.instances == nil {
		c.instances = make(map[string]*instanceCtx)
	}
	c.mu.Unlock()
}

func (c *templateCB) sendTemplateReady() {
	body, _ := proto.Marshal(&pb.DS2DSA_TemplateReady{})
	c.send(pb.DS2DSA_MsgID_TemplateReady, body)
}

func (c *templateCB) send(msgID pb.DS2DSA_MsgID_MsgID, body []byte) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint16(msgID))
	buf.Write(body)
	if conn := c.getConn(); conn != nil {
		_ = conn.SendMsg(buf)
	}
}

func (c *templateCB) OnConnectCB(conn kissnet.Connection) error {
	logrus.Info("template connected to DSA")
	return nil
}

func (c *templateCB) OnDisconnectCB(conn kissnet.Connection) error {
	logrus.Warn("template disconnected from DSA")
	c.setConn(nil)
	// process 模式 DSA 重启而模板仍在：异步重连并重发 TemplateReady（不阻塞 EventLoop）。
	if atomic.LoadInt32(&c.closed) == 0 {
		go c.connectLoop(false)
	}
	return nil
}

func (c *templateCB) OnMsgCB(conn kissnet.Connection, msg []byte) error {
	if len(msg) < 2 {
		return fmt.Errorf("msg len error")
	}
	msgID := pb.DSA2DS_MsgID_MsgID(binary.LittleEndian.Uint16(msg))
	body := msg[2:]
	switch msgID {
	case pb.DSA2DS_MsgID_ForkInstance:
		c.handleFork(body)
	default:
		return fmt.Errorf("unexpected template msgID %d", msgID)
	}
	return nil
}

func (c *templateCB) handleFork(body []byte) {
	c.ensure()
	req := &pb.DSA2DS_ForkInstanceReq{}
	if err := proto.Unmarshal(body, req); err != nil {
		logrus.WithError(err).Error("ForkInstance unmarshal")
		c.sendForkFailed(req.DsID, "unmarshal error")
		return
	}
	logrus.WithFields(logrus.Fields{"dsID": req.DsID, "realmID": req.RealmID, "port": req.Port}).Info("ForkInstance")

	advIP := req.AdvertiseIP
	if advIP == "" {
		advIP = c.advertiseIP
	}
	inst := newInstanceCtx(req.DsID, req.RealmID, int(req.Port), advIP, c.dsAddr, c.event)
	c.mu.Lock()
	c.instances[req.DsID] = inst
	c.mu.Unlock()
	if err := inst.start(); err != nil {
		logrus.WithError(err).WithField("dsID", req.DsID).Error("instance start failed")
		c.sendForkFailed(req.DsID, err.Error())
		c.mu.Lock()
		delete(c.instances, req.DsID)
		c.mu.Unlock()
	}
}

func (c *templateCB) sendForkFailed(dsID, reason string) {
	body, _ := proto.Marshal(&pb.DS2DSA_ForkFailed{DsID: dsID, Reason: reason})
	c.send(pb.DS2DSA_MsgID_ForkFailed, body)
}

func (c *templateCB) close() {
	c.closeOnce.Do(func() {
		atomic.StoreInt32(&c.closed, 1)
		if conn := c.getConn(); conn != nil {
			conn.Close()
		}
		c.mu.Lock()
		for _, inst := range c.instances {
			inst.stop()
		}
		c.mu.Unlock()
	})
}

// instanceCtx 模拟一个 fork 出来的 DS 实例。
type instanceCtx struct {
	dsID     string
	realmID  string
	port     int
	advIP    string
	dsaAddr  string
	event    *kissnet.NetEvent
	conn     kissnet.Connection
	listener net.Listener
	stopOnce sync.Once
}

func newInstanceCtx(dsID, realmID string, port int, advIP, dsaAddr string, event *kissnet.NetEvent) *instanceCtx {
	return &instanceCtx{dsID: dsID, realmID: realmID, port: port, advIP: advIP, dsaAddr: dsaAddr, event: event}
}

func (i *instanceCtx) start() error {
	// 1. 在游戏口上 listen，供玩家连接（联调验证用）。
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", i.port))
	if err != nil {
		// 端口可能被占：报 ForkFailed 让 DSA 释放口并重选。
		return fmt.Errorf("listen game port %d: %w", i.port, err)
	}
	i.listener = ln
	go i.acceptGame()

	// 2. 连上 DSA 实例口，发 CreateOK。
	cb := &instanceCB{inst: i}
	conn, err := kissnet.TCPConnector(i.dsaAddr, cb, i.event)
	if err != nil {
		ln.Close()
		return fmt.Errorf("connect DSA ds port: %w", err)
	}
	i.conn = conn
	cb.conn = conn
	i.sendCreateOK()
	logrus.WithFields(logrus.Fields{"dsID": i.dsID, "dsAddr": i.addr()}).Info("instance CreateOK sent")
	return nil
}

func (i *instanceCtx) acceptGame() {
	for {
		c, err := i.listener.Accept()
		if err != nil {
			return
		}
		logrus.WithFields(logrus.Fields{"dsID": i.dsID, "remote": c.RemoteAddr()}).Info("game client connected")
		// 模拟 DS：保持连接直到客户端断开；不做协议处理。
		go func(c net.Conn) {
			buf := make([]byte, 1024)
			for {
				if _, err := c.Read(buf); err != nil {
					_ = c.Close()
					return
				}
			}
		}(c)
	}
}

func (i *instanceCtx) addr() string {
	return fmt.Sprintf("%s:%d", i.advIP, i.port)
}

func (i *instanceCtx) sendCreateOK() {
	body, _ := proto.Marshal(&pb.DS2DSA_CreateOK{
		DsID:    i.dsID,
		RealmID: i.realmID,
		DsAddr:  i.addr(),
	})
	i.send(pb.DS2DSA_MsgID_DSCreateOK, body)
}

func (i *instanceCtx) send(msgID pb.DS2DSA_MsgID_MsgID, body []byte) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint16(msgID))
	buf.Write(body)
	if i.conn != nil {
		_ = i.conn.SendMsg(buf)
	}
}

func (i *instanceCtx) stop() {
	i.stopOnce.Do(func() {
		if i.conn != nil {
			i.conn.Close()
		}
		if i.listener != nil {
			_ = i.listener.Close()
		}
	})
}

// instanceCB 处理实例口连接（收 StopDS）。
type instanceCB struct {
	inst *instanceCtx
	conn kissnet.Connection
}

func (cb *instanceCB) OnConnectCB(conn kissnet.Connection) error { return nil }

func (cb *instanceCB) OnDisconnectCB(conn kissnet.Connection) error {
	logrus.WithField("dsID", cb.inst.dsID).Info("instance disconnected from DSA")
	cb.inst.stop()
	return nil
}

func (cb *instanceCB) OnMsgCB(conn kissnet.Connection, msg []byte) error {
	if len(msg) < 2 {
		return fmt.Errorf("msg len error")
	}
	msgID := pb.DSA2DS_MsgID_MsgID(binary.LittleEndian.Uint16(msg))
	switch msgID {
	case pb.DSA2DS_MsgID_StopDS:
		req := &pb.DSA2DS_StopDSReq{}
		_ = proto.Unmarshal(msg[2:], req)
		logrus.WithField("dsID", req.DsID).Info("instance StopDS received")
		cb.inst.stop()
	case pb.DSA2DS_MsgID_QueryDsStatus:
		// 模拟：回一个 UpdateState 即可（联调用）。
		cb.inst.send(pb.DS2DSA_MsgID_DSUpdateState, []byte{})
	default:
		return fmt.Errorf("unexpected instance msgID %d", msgID)
	}
	return nil
}

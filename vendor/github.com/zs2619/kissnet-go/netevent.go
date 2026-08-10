package kissnet

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

type msgCallback func()

// NetEvent 业务回调的串行事件循环。
// 关停约定：先 WaitConnectionsClosed（保证 Schedule* 结束），再 Close。
type NetEvent struct {
	queue     chan msgCallback
	ctx       context.Context
	cancel    context.CancelFunc
	closeOnce sync.Once

	timeOutMap   sync.Map
	timeOutIDGen int64
}

func NewNetEvent() *NetEvent {
	ctx, cancel := context.WithCancel(context.Background())
	return &NetEvent{
		queue:  make(chan msgCallback, 1024*64),
		ctx:    ctx,
		cancel: cancel,
	}
}

// ScheduleMsg 非阻塞投递；队列满或已关闭则丢弃。
func (e *NetEvent) ScheduleMsg(cb msgCallback) {
	if cb == nil {
		return
	}
	select {
	case e.queue <- cb:
	case <-e.ctx.Done():
		return
	default:
		logrus.Warn("NetEvent queue full, drop msg")
	}
}

// ScheduleMsgTimeout 在 timeout 内阻塞投递；超时或已关闭则丢弃。
func (e *NetEvent) ScheduleMsgTimeout(cb msgCallback, timeout time.Duration) bool {
	if cb == nil {
		return false
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case e.queue <- cb:
		return true
	case <-e.ctx.Done():
		return false
	case <-timer.C:
		logrus.WithField("timeout", timeout).Warn("NetEvent enqueue timeout, drop callback")
		return false
	}
}

// EventLoop 阻塞执行回调，直到 Close；关闭后排空已入队回调再退出。
func (e *NetEvent) EventLoop() {
	logrus.Info("Start Main EventLoop")
	for {
		select {
		case cb := <-e.queue:
			cb()
		case <-e.ctx.Done():
			e.drainAndQuit()
			return
		}
	}
}

func (e *NetEvent) drainAndQuit() {
	for {
		select {
		case cb := <-e.queue:
			cb()
		default:
			logrus.Info("Main MsgLoop Quit (drained)")
			return
		}
	}
}

// Close 停止 EventLoop 并取消定时器；幂等。
func (e *NetEvent) Close() {
	e.closeOnce.Do(func() {
		e.cancel()
		e.timeOutMap.Range(func(key, value interface{}) bool {
			to := value.(*TimeOut)
			to.quitOnce.Do(func() { close(to.quit) })
			e.timeOutMap.Delete(key)
			return true
		})
	})
}

type TimeOut struct {
	msInterval time.Duration
	cb         func() error
	quit       chan struct{}
	quitOnce   sync.Once
	ne         *NetEvent
}

func (t *TimeOut) Run() {
	ticker := time.NewTicker(t.msInterval)
	defer ticker.Stop()
	for {
		select {
		case <-t.quit:
			return
		case <-ticker.C:
			cb := t.cb
			t.ne.ScheduleMsg(func() {
				if err := cb(); err != nil {
					logrus.WithError(err).Error("timer cb error")
				}
			})
		}
	}
}

// ScheduleTimer 启动周期定时器，回调投递到 EventLoop；已关闭返回 0。
func (e *NetEvent) ScheduleTimer(ms int64, cb func() error) int64 {
	if e.ctx.Err() != nil {
		return 0
	}
	timeOut := &TimeOut{
		msInterval: time.Duration(ms) * time.Millisecond,
		cb:         cb,
		quit:       make(chan struct{}),
		ne:         e,
	}
	id := atomic.AddInt64(&e.timeOutIDGen, 1)
	e.timeOutMap.Store(id, timeOut)
	go timeOut.Run()
	return id
}

// CancelTimer 取消定时器。
func (e *NetEvent) CancelTimer(id int64) {
	v, ok := e.timeOutMap.LoadAndDelete(id)
	if !ok {
		return
	}
	to := v.(*TimeOut)
	to.quitOnce.Do(func() { close(to.quit) })
}

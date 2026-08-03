package kissnet

import (
	"sync"
	"sync/atomic"
	"time"
)

// ConnectionMgr 管理活跃连接。
// CloseAll 后拒绝 add；WaitEmpty 等待所有已登记连接的 start 退出。
type ConnectionMgr struct {
	sesMgr   sync.Map
	sesIDGen int64
	count    atomic.Int64

	wg     sync.WaitGroup
	mu     sync.Mutex
	closed bool
}

// add 登记连接并分配 id；已关停时返回 false。
func (m *ConnectionMgr) add(c Connection) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return false
	}
	id := atomic.AddInt64(&m.sesIDGen, 1)
	c.setID(id)
	m.sesMgr.Store(id, c)
	m.count.Add(1)
	m.wg.Add(1)
	return true
}

func (m *ConnectionMgr) del(c Connection) {
	if c == nil {
		return
	}
	if _, ok := m.sesMgr.LoadAndDelete(c.getID()); !ok {
		return
	}
	m.count.Add(-1)
	m.wg.Done()
}

func (m *ConnectionMgr) Count() int64 {
	return m.count.Load()
}

// WaitEmpty 等待所有连接退出，超时返回 false。
func (m *ConnectionMgr) WaitEmpty(timeout time.Duration) bool {
	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return true
	case <-time.After(timeout):
		return false
	}
}

// CloseAll 拒绝后续 add，并关闭当前连接。
func (m *ConnectionMgr) CloseAll() {
	m.mu.Lock()
	m.closed = true
	m.mu.Unlock()

	m.sesMgr.Range(func(_, v interface{}) bool {
		v.(Connection).Close()
		return true
	})
}

package gameplayer

import "github.com/zs2619/kissnet-go"

// netEvent 主事件循环；player 状态变更须投递到此串行执行。
var netEvent *kissnet.NetEvent

// SetNetEvent 由 main 在启动 EventLoop 前注入。
func SetNetEvent(ne *kissnet.NetEvent) {
	netEvent = ne
}

// schedule 将回调投递到 EventLoop；尚未注入时同步执行（仅启动极早期）。
func schedule(fn func()) {
	if fn == nil {
		return
	}
	if netEvent != nil {
		netEvent.ScheduleMsg(fn)
		return
	}
	fn()
}

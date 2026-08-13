// Package safego 提供带 panic 保护的 goroutine 启动能力。
//
// 背景：Go 中任意 goroutine 未 recover 的 panic 会导致整个进程崩溃，
// 且其他 goroutine 的 recover 无法捕获。后台任务（异步回调、定时任务、
// MQ 消费、Pub/Sub 订阅等）必须在自身内部做好 panic 防护。
package safego

import "log/slog"

// Go 在独立 goroutine 中执行 fn，并捕获 panic 避免进程崩溃。
// name 用于日志标识后台任务，便于排查。
func Go(name string, fn func()) {
	go func() {
		defer Recover(name)
		fn()
	}()
}

// Recover 捕获 panic 并记录日志，配合 defer 使用。
// 常用于常驻循环中保护单次执行，panic 后循环仍可继续运行。
func Recover(name string) {
	if r := recover(); r != nil {
		slog.Error("后台任务 panic", "task", name, "panic", r)
	}
}

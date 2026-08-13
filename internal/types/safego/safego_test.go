package safego

import (
	"sync"
	"testing"
)

// TestGo_RecoverPanic 验证 Go 启动的 goroutine panic 会被捕获，不会导致进程崩溃。
// 若 Recover 失效，测试进程将直接 panic 崩溃。
func TestGo_RecoverPanic(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	Go("test_panic", func() {
		defer wg.Done()
		panic("boom")
	})
	wg.Wait()
}

// TestGo_ExecuteNormally 验证正常函数照常执行。
func TestGo_ExecuteNormally(t *testing.T) {
	done := make(chan struct{})
	Go("test_normal", func() {
		close(done)
	})
	<-done
}

// TestRecover_ReturnsTrue 验证 Recover 在 panic 时能捕获并返回。
func TestRecover_NoPanic(t *testing.T) {
	func() {
		defer Recover("test_no_panic")
		// 无 panic，不应崩溃
	}()
}

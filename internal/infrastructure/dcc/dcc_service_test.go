package dcc

import (
	"context"
	"sync"
	"testing"
	"time"
)

type memBus struct {
	mu   sync.Mutex
	subs map[string][]func(string)
}

func newMemBus() *memBus {
	return &memBus{subs: make(map[string][]func(string))}
}

func (b *memBus) Publish(ctx context.Context, channel string, message string) error {
	b.mu.Lock()
	handlers := append([]func(string){}, b.subs[channel]...)
	b.mu.Unlock()
	for _, h := range handlers {
		h(message)
	}
	return nil
}

func (b *memBus) Subscribe(ctx context.Context, channel string, handler func(payload string)) error {
	b.mu.Lock()
	b.subs[channel] = append(b.subs[channel], handler)
	b.mu.Unlock()
	return nil
}

func TestUpdateLocal(t *testing.T) {
	s := New("0", "100", "s02c02", "0")
	if s.IsDowngradeSwitch() {
		t.Fatal("default not downgrade")
	}
	s.Update("downgradeSwitch", "1")
	if !s.IsDowngradeSwitch() {
		t.Fatal("should downgrade")
	}
	s.Update("cutRange", "0")
	// hash 任意用户 lastTwo 可能 >0
	_ = s.IsCutRange("xfg01")
	if !s.IsSCBlackIntercept("s02", "c02") {
		t.Fatal("blacklist")
	}
	if s.IsSCBlackIntercept("s01", "c01") {
		t.Fatal("should pass")
	}
}

func TestCrossInstanceBroadcast(t *testing.T) {
	bus := newMemBus()
	a := New("0", "100", "", "0")
	b := New("0", "100", "", "0")
	a.EnableBroadcast(bus, bus, DefaultChannel, "inst-a")
	b.EnableBroadcast(bus, bus, DefaultChannel, "inst-b")

	// 等订阅注册
	time.Sleep(20 * time.Millisecond)

	a.Update("downgradeSwitch", "1")
	// 同步 bus 直接调用 handler
	time.Sleep(20 * time.Millisecond)

	if !b.IsDowngradeSwitch() {
		t.Fatal("instance B should receive downgrade from A")
	}
	snap := b.Snapshot()
	if snap["downgradeSwitch"] != "1" {
		t.Fatalf("snap %v", snap)
	}
}

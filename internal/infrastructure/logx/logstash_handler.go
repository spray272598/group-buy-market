package logx

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"os"
	"sync"
	"time"
)

// LogstashHandler 将 slog 以 JSON Lines 发到 Logstash TCP:4560（ELK）
// 连接失败不阻断主流程，仅丢弃并定期重试。
type LogstashHandler struct {
	inner  slog.Handler
	addr   string
	mu     sync.Mutex
	conn   net.Conn
	lastTry time.Time
}

func NewLogstashHandler(inner slog.Handler, addr string) *LogstashHandler {
	if addr == "" {
		addr = "127.0.0.1:4560"
	}
	return &LogstashHandler{inner: inner, addr: addr}
}

func (h *LogstashHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *LogstashHandler) Handle(ctx context.Context, r slog.Record) error {
	// 先写本地
	_ = h.inner.Handle(ctx, r)

	attrs := map[string]any{
		"time":    r.Time.Format(time.RFC3339Nano),
		"level":   r.Level.String(),
		"msg":     r.Message,
		"service": "group-buy-market",
	}
	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.Any()
		return true
	})
	b, err := json.Marshal(attrs)
	if err != nil {
		return nil
	}
	b = append(b, '\n')
	h.write(b)
	return nil
}

func (h *LogstashHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &LogstashHandler{inner: h.inner.WithAttrs(attrs), addr: h.addr}
}

func (h *LogstashHandler) WithGroup(name string) slog.Handler {
	return &LogstashHandler{inner: h.inner.WithGroup(name), addr: h.addr}
}

func (h *LogstashHandler) write(b []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.conn == nil {
		if time.Since(h.lastTry) < 5*time.Second {
			return
		}
		h.lastTry = time.Now()
		c, err := net.DialTimeout("tcp", h.addr, 2*time.Second)
		if err != nil {
			return
		}
		h.conn = c
	}
	_ = h.conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
	if _, err := h.conn.Write(b); err != nil {
		_ = h.conn.Close()
		h.conn = nil
	}
}

// SetupDefault 若环境变量 GBM_LOGSTASH_ADDR 存在则启用双写
func SetupDefault() {
	addr := os.Getenv("GBM_LOGSTASH_ADDR")
	base := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	if addr == "" {
		slog.SetDefault(slog.New(base))
		return
	}
	slog.SetDefault(slog.New(NewLogstashHandler(base, addr)))
}

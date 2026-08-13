// Package trace 提供请求链路追踪 traceId 的 context 传递能力。
//
// 对齐 Java 端 TraceIdFilter + MDC 语义：
//   - 每个入口请求分配唯一 traceId
//   - 通过 context.Context 在 goroutine / RPC / MQ 调用间透传
//   - slog 日志自动携带 trace_id 字段，ELK 可按 traceId 全链路检索
package trace

import (
	"context"
	"crypto/rand"
	"fmt"
)

// traceKey 是 context.Value 的私有类型，防止碰撞
type traceKey struct{}

// FromContext 从 context 中提取 traceId，空字符串表示未设置
func FromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v, ok := ctx.Value(traceKey{}).(string); ok {
		return v
	}
	return ""
}

// SetContext 将 traceId 注入 context，返回新的 context
func SetContext(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceKey{}, traceID)
}

// Generate 生成新的随机 traceId（16 hex 字符）
func Generate() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}

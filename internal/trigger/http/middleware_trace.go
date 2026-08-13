package http

import (
	"crypto/rand"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"group-buy-market/internal/infrastructure/trace"
)

// TraceIDHeader 请求头中传递 traceId 的 key（对齐 Java TraceIdFilter）
const TraceIDHeader = "X-Trace-ID"

// TraceMiddleware 为每个请求生成/提取 traceId 并注入 context，
// 同时写入响应头以便前端排查问题。
// 对齐 Java com.bugstack.trigger.filter.TraceIdFilter
func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader(TraceIDHeader)
		if traceID == "" {
			traceID = generateTraceID()
		}
		ctx := trace.SetContext(c.Request.Context(), traceID)
		c.Request = c.Request.WithContext(ctx)

		// 响应头回写，方便前端/网关串联
		c.Header(TraceIDHeader, traceID)

		c.Next()
	}
}

// generateTraceID 生成 16 字符随机 traceId（类似 Java UUID 截断）
func generateTraceID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// GetTraceID 从 gin.Context 提取当前请求的 traceId
func GetTraceID(c *gin.Context) string {
	return trace.FromContext(c.Request.Context())
}

// TraceKeyFunc 从请求中提取 userId，附加 traceId 用于日志关联
func TraceKeyFunc(keyFunc func(*gin.Context) string) func(*gin.Context) string {
	return func(c *gin.Context) string {
		key := keyFunc(c)
		if key == "" {
			return ""
		}
		// 追加 traceId 方便日志检索：userId@traceId
		tid := GetTraceID(c)
		if tid != "" {
			return key + "@" + tid
		}
		return key
	}
}

// extractUserIDFromQuery 从 query 参数提取 userId（用于 GET /api/market/index）
func ExtractUserIDFromQuery(c *gin.Context) string {
	return c.Query("userId")
}

// extractUserIDFromBody 从 JSON body 提取 userId（用于 POST 接口）
type userIDBody struct {
	UserID string `json:"userId"`
}

func ExtractUserIDFromBody(c *gin.Context) string {
	var body userIDBody
	if err := c.ShouldBindJSON(&body); err != nil {
		return ""
	}
	// 缓存回 c，避免下游再次 ShouldBind 失败
	c.Set("cached_user_id_body", body)
	return body.UserID
}

// CachedUserIDBody 获取缓存的 body 解析结果（配合 ExtractUserIDFromBody 使用）
func CachedUserIDBody(c *gin.Context) (userIDBody, bool) {
	v, ok := c.Get("cached_user_id_body")
	if !ok {
		return userIDBody{}, false
	}
	return v.(userIDBody), true
}

// isHealthOrStatic 判断是否为健康检查或静态资源路径（跳过 trace 注入）
func isHealthOrStatic(path string) bool {
	return path == "/health" ||
		strings.HasPrefix(path, "/metrics") ||
		strings.HasPrefix(path, "/test/") ||
		strings.HasPrefix(path, "/swagger") ||
		path == "/openapi.yaml" ||
		path == "/"
}

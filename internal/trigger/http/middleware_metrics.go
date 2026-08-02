package http

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"group-buy-market/internal/infrastructure/metrics"
)

// metricsBodyWriter 捕获响应体以便解析业务 code（仅小体积 JSON API）
type metricsBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *metricsBodyWriter) Write(b []byte) (int, error) {
	if w.body != nil && w.body.Len()+len(b) < 64*1024 {
		_, _ = w.body.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

func (w *metricsBodyWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

// MetricsMiddleware 全链路 HTTP 指标：QPS / 延迟 / 状态码 / 业务码
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// /metrics 自身不做业务解析，避免干扰 scrape
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()
		metrics.HTTPInFlight.Inc()
		defer metrics.HTTPInFlight.Dec()

		bw := &metricsBodyWriter{ResponseWriter: c.Writer, body: &bytes.Buffer{}}
		c.Writer = bw

		c.Next()

		path := c.FullPath()
		if path == "" {
			path = "unmatched"
		}
		method := c.Request.Method
		status := strconv.Itoa(c.Writer.Status())
		metrics.ObserveHTTP(method, path, status, time.Since(start).Seconds())

		// 业务 API：解析统一 Response.code
		if strings.HasPrefix(path, "/api/") && bw.body.Len() > 0 {
			if code := extractBizCode(bw.body.Bytes()); code != "" {
				metrics.ObserveBiz(path, code)
			}
		}
	}
}

type bizCodeProbe struct {
	Code string `json:"code"`
}

func extractBizCode(body []byte) string {
	// 快速路径：非 JSON 对象跳过
	trim := bytes.TrimSpace(body)
	if len(trim) == 0 || trim[0] != '{' {
		return ""
	}
	var p bizCodeProbe
	if err := json.Unmarshal(trim, &p); err != nil {
		return ""
	}
	return p.Code
}

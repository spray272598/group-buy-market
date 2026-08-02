package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"group-buy-market/internal/infrastructure/metrics"
)

func TestExtractBizCode(t *testing.T) {
	if got := extractBizCode([]byte(`{"code":"0000","info":"成功"}`)); got != "0000" {
		t.Fatalf("want 0000 got %q", got)
	}
	if got := extractBizCode([]byte(`{"code":"0006","info":"接口限流"}`)); got != "0006" {
		t.Fatalf("want 0006 got %q", got)
	}
	if got := extractBizCode([]byte(`not-json`)); got != "" {
		t.Fatalf("non-json should be empty, got %q", got)
	}
	if got := extractBizCode([]byte(``)); got != "" {
		t.Fatalf("empty should be empty")
	}
}

func TestMetricsMiddlewareRecordsHTTPAndBiz(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(MetricsMiddleware())
	r.POST("/api/v1/gbm/index/query_group_buy_market_config", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": "0000", "info": "成功"})
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/gbm/index/query_group_buy_market_config", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("status %d", w.Code)
	}

	if n := testutil.CollectAndCount(metrics.HTTPRequests); n < 1 {
		t.Fatalf("expected http metrics series, got %d", n)
	}
	if n := testutil.CollectAndCount(metrics.BizRequests); n < 1 {
		t.Fatalf("expected biz metrics series, got %d", n)
	}
	if n := testutil.CollectAndCount(metrics.HTTPDuration); n < 1 {
		t.Fatalf("expected duration histogram series, got %d", n)
	}
}

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"group-buy-market/internal/infrastructure/trace"
)

func TestTraceMiddleware_GeneratesTraceID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(TraceMiddleware())
	engine.GET("/test", func(c *gin.Context) {
		tid := GetTraceID(c)
		if tid == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "no trace id"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"trace_id": tid})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	tid := w.Header().Get(TraceIDHeader)
	if tid == "" {
		t.Fatal("response missing X-Trace-ID header")
	}
	if len(tid) < 8 {
		t.Fatalf("trace_id too short: %q", tid)
	}
}

func TestTraceMiddleware_PropagatesExisting(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(TraceMiddleware())
	engine.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"trace_id": GetTraceID(c)})
	})

	existing := "a1b2c3d4e5f6"
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(TraceIDHeader, existing)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if got := w.Header().Get(TraceIDHeader); got != existing {
		t.Errorf("expected trace_id %q, got %q", existing, got)
	}
}

func TestTraceContext_RoundTrip(t *testing.T) {
	ctx := trace.SetContext(context.Background(), trace.Generate())
	if tid := trace.FromContext(ctx); tid == "" {
		t.Fatal("FromContext returned empty after SetContext")
	}
	if trace.FromContext(nil) != "" {
		t.Fatal("FromContext(nil) should be empty")
	}
}

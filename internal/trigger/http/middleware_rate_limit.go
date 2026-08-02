package http

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"group-buy-market/internal/types/enums"
	"group-buy-market/internal/types/response"
)

// RateLimiter 简易令牌桶限流（对齐 Java @RateLimiterAccessInterceptor 语义）
// 按 userId 维度限制 QPS，超限返回 RATE_LIMITER
type userLimiter struct {
	lim      *rate.Limiter
	lastSeen time.Time
}

type RateLimitStore struct {
	mu       sync.Mutex
	users    map[string]*userLimiter
	r        rate.Limit
	burst    int
	enabled  bool
}

func NewRateLimitStore(qps float64, burst int) *RateLimitStore {
	if burst <= 0 {
		burst = 1
	}
	return &RateLimitStore{
		users:   make(map[string]*userLimiter),
		r:       rate.Limit(qps),
		burst:   burst,
		enabled: qps > 0,
	}
}

func (s *RateLimitStore) allow(userID string) bool {
	if !s.enabled || userID == "" {
		return true
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.users[userID]
	if !ok {
		u = &userLimiter{lim: rate.NewLimiter(s.r, s.burst), lastSeen: time.Now()}
		s.users[userID] = u
	}
	u.lastSeen = time.Now()
	return u.lim.Allow()
}

// RateLimitByUserJSON 从 JSON body 提取 userId 做限流（用于 index 接口）
// 注意：会缓存 body，需配合 ShouldBind 使用前先读
func RateLimitMiddleware(store *RateLimitStore, keyFunc func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := keyFunc(c)
		if key != "" && !store.allow(key) {
			c.AbortWithStatusJSON(http.StatusOK, response.Fail[any](enums.RATE_LIMITER.Code, enums.RATE_LIMITER.Info))
			return
		}
		c.Next()
	}
}

// Package metrics 全链路可观测指标（Prometheus）
//
// 分层：
//   - HTTP / 业务码 / 限流（入口）
//   - 库存 / DCC / Job / Notify（业务与异步）
//   - Redis / MQ 客户端操作（依赖调用）
//   - DB 连接池 + 依赖健康探针（基础设施）
//
// 命名统一 gbm_ 前缀，便于 Grafana 过滤。
package metrics

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// ---------- HTTP / 入口 ----------
	HTTPRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gbm_http_requests_total",
		Help: "Total HTTP requests by method, path and status code",
	}, []string{"method", "path", "status"})

	HTTPDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "gbm_http_request_duration_seconds",
		Help:    "HTTP request latency in seconds",
		Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	}, []string{"method", "path"})

	HTTPInFlight = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "gbm_http_requests_in_flight",
		Help: "Number of HTTP requests currently being served",
	})

	RateLimit = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gbm_rate_limit_total",
		Help: "Rate limit decisions by result (allow|deny)",
	}, []string{"result"})

	BizRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gbm_biz_requests_total",
		Help: "Business API responses by path and business code",
	}, []string{"path", "code"})

	// ---------- 业务：库存 / DCC ----------
	StockOps = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gbm_stock_ops_total",
		Help: "Team stock operations (occupy|recovery|refund_recovery)",
	}, []string{"op", "result"}) // result: success|fail|skip|error

	DCCUpdates = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gbm_dcc_updates_total",
		Help: "DCC config updates by key and source (local|broadcast)",
	}, []string{"key", "source"})

	// ---------- 异步：Job / Notify / MQ ----------
	JobRuns = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gbm_job_runs_total",
		Help: "Background job executions by job name and result",
	}, []string{"job", "result"})

	JobProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gbm_job_processed_total",
		Help: "Items processed by background jobs",
	}, []string{"job", "result"})

	JobDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "gbm_job_duration_seconds",
		Help:    "Background job execution duration",
		Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 15, 30, 60},
	}, []string{"job"})

	Notify = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gbm_notify_total",
		Help: "Outbound group-buy notify attempts by channel and result",
	}, []string{"channel", "result"})

	NotifyDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "gbm_notify_duration_seconds",
		Help:    "Outbound notify latency (http|mq)",
		Buckets: []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	}, []string{"channel"})

	MQPublish = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gbm_mq_publish_total",
		Help: "RabbitMQ publish attempts by routing_key and result",
	}, []string{"routing_key", "result"})

	MQConsume = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gbm_mq_consume_total",
		Help: "RabbitMQ consume handler results by queue and result",
	}, []string{"queue", "result"})

	MQConsumeDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "gbm_mq_consume_duration_seconds",
		Help:    "RabbitMQ consume handler duration",
		Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 15},
	}, []string{"queue"})

	// ---------- 依赖客户端：Redis ----------
	RedisOps = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gbm_redis_ops_total",
		Help: "Redis client operations by op and result",
	}, []string{"op", "result"})

	RedisDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "gbm_redis_op_duration_seconds",
		Help:    "Redis client operation latency",
		Buckets: []float64{0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
	}, []string{"op"})

	// ---------- 依赖健康 + DB 连接池 ----------
	// 1=up 0=down
	DependencyUp = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gbm_dependency_up",
		Help: "Dependency health (1=up, 0=down): mysql|redis|rabbitmq",
	}, []string{"name"})

	DBOpenConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "gbm_db_open_connections",
		Help: "Database open connections",
	})
	DBInUse = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "gbm_db_in_use",
		Help: "Database connections currently in use",
	})
	DBIdle = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "gbm_db_idle",
		Help: "Database idle connections",
	})
	DBWaitCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "gbm_db_wait_count",
		Help: "Total number of connections waited for (cumulative from stats)",
	})
	DBMaxOpen = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "gbm_db_max_open_connections",
		Help: "Configured max open connections",
	})
)

// ---------- 便捷观察函数 ----------

func ObserveHTTP(method, path, status string, seconds float64) {
	HTTPRequests.WithLabelValues(method, path, status).Inc()
	HTTPDuration.WithLabelValues(method, path).Observe(seconds)
}

func ObserveBiz(path, code string) {
	if path == "" || code == "" {
		return
	}
	BizRequests.WithLabelValues(path, code).Inc()
}

func ObserveRateLimit(allowed bool) {
	if allowed {
		RateLimit.WithLabelValues("allow").Inc()
	} else {
		RateLimit.WithLabelValues("deny").Inc()
	}
}

func ObserveJob(job, result string) {
	JobRuns.WithLabelValues(job, result).Inc()
}

func ObserveJobProcessed(job, result string, n int) {
	if n <= 0 {
		return
	}
	JobProcessed.WithLabelValues(job, result).Add(float64(n))
}

func ObserveJobDuration(job string, seconds float64) {
	JobDuration.WithLabelValues(job).Observe(seconds)
}

func ObserveNotify(channel, result string) {
	Notify.WithLabelValues(channel, result).Inc()
}

func ObserveNotifyDuration(channel string, seconds float64) {
	NotifyDuration.WithLabelValues(channel).Observe(seconds)
}

func ObserveStock(op, result string) {
	StockOps.WithLabelValues(op, result).Inc()
}

func ObserveDCC(key, source string) {
	DCCUpdates.WithLabelValues(key, source).Inc()
}

func ObserveMQPublish(routingKey, result string) {
	if routingKey == "" {
		routingKey = "unknown"
	}
	MQPublish.WithLabelValues(routingKey, result).Inc()
}

func ObserveMQConsume(queue, result string, seconds float64) {
	MQConsume.WithLabelValues(queue, result).Inc()
	MQConsumeDuration.WithLabelValues(queue).Observe(seconds)
}

func ObserveRedis(op string, err error, seconds float64) {
	result := "ok"
	if err != nil {
		result = "error"
	}
	RedisOps.WithLabelValues(op, result).Inc()
	RedisDuration.WithLabelValues(op).Observe(seconds)
}

// ---------- 采集器：DB 池 + 依赖健康 ----------

// DepChecker 依赖健康检查函数（返回 nil 表示 up）
type DepChecker func(ctx context.Context) error

// Collectors 后台采集句柄，Stop 可关闭
type Collectors struct {
	stop chan struct{}
	once sync.Once
}

// StartCollectors 启动 DB 连接池与依赖健康后台采集
func StartCollectors(db *sql.DB, deps map[string]DepChecker) *Collectors {
	c := &Collectors{stop: make(chan struct{})}
	go c.loop(db, deps)
	return c
}

func (c *Collectors) Stop() {
	c.once.Do(func() { close(c.stop) })
}

func (c *Collectors) loop(db *sql.DB, deps map[string]DepChecker) {
	// 启动即采一次
	c.sample(db, deps)
	t := time.NewTicker(10 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-c.stop:
			return
		case <-t.C:
			c.sample(db, deps)
		}
	}
}

func (c *Collectors) sample(db *sql.DB, deps map[string]DepChecker) {
	if db != nil {
		s := db.Stats()
		DBOpenConnections.Set(float64(s.OpenConnections))
		DBInUse.Set(float64(s.InUse))
		DBIdle.Set(float64(s.Idle))
		DBWaitCount.Set(float64(s.WaitCount))
		DBMaxOpen.Set(float64(s.MaxOpenConnections))
	}
	for name, check := range deps {
		if check == nil {
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err := check(ctx)
		cancel()
		if err != nil {
			DependencyUp.WithLabelValues(name).Set(0)
		} else {
			DependencyUp.WithLabelValues(name).Set(1)
		}
	}
}

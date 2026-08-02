package metrics

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestObserveHelpers(t *testing.T) {
	ObserveHTTP("GET", "/health", "200", 0.01)
	ObserveBiz("/api/v1/x", "0000")
	ObserveRateLimit(true)
	ObserveRateLimit(false)
	ObserveStock("occupy", "success")
	ObserveDCC("downgradeSwitch", "local")
	ObserveJob("notify", "success")
	ObserveJobProcessed("notify", "wait", 2)
	ObserveJobDuration("notify", 0.05)
	ObserveNotify("http", "success")
	ObserveNotifyDuration("http", 0.02)
	ObserveMQPublish("topic.team_success", "success")
	ObserveMQConsume("q1", "success", 0.01)
	ObserveRedis("ping", nil, 0.001)
	ObserveRedis("get", errors.New("x"), 0.002)

	if n := testutil.CollectAndCount(HTTPRequests); n < 1 {
		t.Fatalf("http series %d", n)
	}
	if n := testutil.CollectAndCount(BizRequests); n < 1 {
		t.Fatalf("biz series %d", n)
	}
	if n := testutil.CollectAndCount(StockOps); n < 1 {
		t.Fatalf("stock series %d", n)
	}
	if n := testutil.CollectAndCount(RedisOps); n < 1 {
		t.Fatalf("redis series %d", n)
	}
}

func TestCollectorsDependency(t *testing.T) {
	c := StartCollectors(nil, map[string]DepChecker{
		"redis": func(ctx context.Context) error { return nil },
		"mysql": func(ctx context.Context) error { return errors.New("down") },
	})
	defer c.Stop()
	// 启动即 sample 一次；再等一轮确保 gauge 已写
	time.Sleep(50 * time.Millisecond)
	// 仅验证指标族可采集
	if n := testutil.CollectAndCount(DependencyUp); n < 1 {
		t.Fatalf("dependency_up series %d", n)
	}
}

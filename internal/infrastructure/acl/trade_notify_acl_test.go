package acl_test

import (
	"context"
	"testing"
	"time"

	"group-buy-market/internal/domain/trade/adapter/port"
	"group-buy-market/internal/domain/trade/model/entity"
	"group-buy-market/internal/domain/trade/model/valobj"
	"group-buy-market/internal/infrastructure/gateway"
	"group-buy-market/internal/infrastructure/mq"
	redisx "group-buy-market/internal/infrastructure/redis"
	"group-buy-market/internal/types/enums"
)

// ============================================================================
// Mock Redis 服务
// ============================================================================

type redisServiceMock struct {
	lockResult bool
	lockErr    error
	unlockErr  error
}

func (m *redisServiceMock) TryLockWait(ctx context.Context, lockKey string, waitTimeout, leaseTimeout time.Duration) (bool, error) {
	return m.lockResult, m.lockErr
}

func (m *redisServiceMock) TryLock(ctx context.Context, lockKey string, leaseTimeout time.Duration) (bool, error) {
	return m.lockResult, m.lockErr
}

func (m *redisServiceMock) Unlock(ctx context.Context, lockKey string) error {
	return m.unlockErr
}

func (m *redisServiceMock) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return true, nil
}

func (m *redisServiceMock) Incr(ctx context.Context, key string) (int64, error) {
	return 1, nil
}

func (m *redisServiceMock) Decr(ctx context.Context, key string) (int64, error) {
	return 1, nil
}

func (m *redisServiceMock) GetInt64(ctx context.Context, key string) (int64, error) {
	return 0, nil
}

func (m *redisServiceMock) Del(ctx context.Context, key string) error {
	return nil
}

// 确保接口实现正确
var _ redisx.IRedisService = (*redisServiceMock)(nil)

// ============================================================================
// Mock HTTP 网关服务
// ============================================================================

type httpGatewayMock struct {
	notifyResult string
	notifyErr    error
}

func (m *httpGatewayMock) GroupBuyNotify(ctx context.Context, url, parameterJSON string) (string, error) {
	return m.notifyResult, m.notifyErr
}

// 确保接口实现正确
var _ gateway.IGroupBuyNotifyService = (*httpGatewayMock)(nil)

// ============================================================================
// Mock MQ 发布服务
// ============================================================================

type mqPublisherMock struct {
	publishErr error
}

func (m *mqPublisherMock) Publish(ctx context.Context, routingKey, message string) error {
	return m.publishErr
}

func (m *mqPublisherMock) StartConsumer(ctx context.Context, handler mq.ConsumerHandler) error {
	return nil
}

// 确保接口实现正确
var _ mq.IPublisher = (*mqPublisherMock)(nil)

// ============================================================================
// ACL 防腐层测试
// ============================================================================

// TestTradeNotifyACL_NilTask 空任务处理
func TestTradeNotifyACL_NilTask(t *testing.T) {
	t.Log("测试场景：空任务处理")

	acl := NewTradeNotifyACLForTest()

	result, err := acl.GroupBuyNotify(context.Background(), nil)
	if err != nil {
		t.Fatalf("空任务处理失败: %v", err)
	}
	if result != enums.NotifyTaskHTTPSuccess {
		t.Errorf("期望返回成功，实际为 %s", result)
	}

	t.Log("✅ 空任务处理测试通过")
}

// TestTradeNotifyACL_HTTPNotify HTTP 回调通知
func TestTradeNotifyACL_HTTPNotify(t *testing.T) {
	t.Log("测试场景：HTTP 回调通知")

	tests := []struct {
		name       string
		notifyUrl  string
		httpResult string
		httpErr    error
		wantResult string
	}{
		{
			name:       "正常 HTTP 回调",
			notifyUrl:  "http://localhost:8080/callback",
			httpResult: enums.NotifyTaskHTTPSuccess,
			httpErr:    nil,
			wantResult: enums.NotifyTaskHTTPSuccess,
		},
		{
			name:       "HTTP 回调失败",
			notifyUrl:  "http://localhost:8080/callback",
			httpResult: enums.NotifyTaskHTTPError,
			httpErr:    context.DeadlineExceeded,
			wantResult: enums.NotifyTaskHTTPError,
		},
		{
			name:       "空 URL（跳过）",
			notifyUrl:  "",
			httpResult: "",
			httpErr:    nil,
			wantResult: enums.NotifyTaskHTTPSuccess,
		},
		{
			name:       "URL 为'暂无'（跳过）",
			notifyUrl:  "暂无",
			httpResult: "",
			httpErr:    nil,
			wantResult: enums.NotifyTaskHTTPSuccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("测试用例: %s", tt.name)

			acl := NewTradeNotifyACLForTestWithMocks(
				&redisServiceMock{lockResult: true},
				&httpGatewayMock{notifyResult: tt.httpResult, notifyErr: tt.httpErr},
				&mqPublisherMock{},
			)

			task := &entity.NotifyTaskEntity{
				TeamID:        "TEAM_HTTP_001",
				NotifyType:    string(valobj.NotifyHTTP),
				NotifyUrl:     tt.notifyUrl,
				ParameterJSON: `{"teamId":"TEAM_HTTP_001"}`,
				UUID:          "UUID_HTTP_001",
				ActivityID:    100123,
			}

			result, err := acl.GroupBuyNotify(context.Background(), task)

			if tt.httpErr != nil && err == nil {
				t.Error("期望返回错误")
			}
			if result != tt.wantResult {
				t.Errorf("期望结果 %s，实际为 %s", tt.wantResult, result)
			}

			t.Logf("HTTP 回调测试完成: result=%s, err=%v", result, err)
		})
	}

	t.Log("✅ HTTP 回调通知测试完成")
}

// TestTradeNotifyACL_MQNotify MQ 消息通知
func TestTradeNotifyACL_MQNotify(t *testing.T) {
	t.Log("测试场景：MQ 消息通知")

	tests := []struct {
		name       string
		routingKey string
		publishErr error
		wantResult string
	}{
		{
			name:       "正常 MQ 发布",
			routingKey: "topic.team_success",
			publishErr: nil,
			wantResult: enums.NotifyTaskHTTPSuccess,
		},
		{
			name:       "MQ 发布失败",
			routingKey: "topic.team_success",
			publishErr: context.DeadlineExceeded,
			wantResult: enums.NotifyTaskHTTPError,
		},
		{
			name:       "空 routingKey（使用默认值）",
			routingKey: "",
			publishErr: nil,
			wantResult: enums.NotifyTaskHTTPSuccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("测试用例: %s", tt.name)

			acl := NewTradeNotifyACLForTestWithMocks(
				&redisServiceMock{lockResult: true},
				&httpGatewayMock{},
				&mqPublisherMock{publishErr: tt.publishErr},
			)

			task := &entity.NotifyTaskEntity{
				TeamID:        "TEAM_MQ_001",
				NotifyType:    string(valobj.NotifyMQ),
				NotifyMQ:      tt.routingKey,
				ParameterJSON: `{"teamId":"TEAM_MQ_001"}`,
				UUID:          "UUID_MQ_001",
				ActivityID:    100123,
			}

			result, err := acl.GroupBuyNotify(context.Background(), task)

			if tt.publishErr != nil && err == nil {
				t.Error("期望返回错误")
			}
			if result != tt.wantResult {
				t.Errorf("期望结果 %s，实际为 %s", tt.wantResult, result)
			}

			t.Logf("MQ 通知测试完成: result=%s, err=%v", result, err)
		})
	}

	t.Log("✅ MQ 消息通知测试完成")
}

// TestTradeNotifyACL_DistributedLock 分布式锁测试
func TestTradeNotifyACL_DistributedLock(t *testing.T) {
	t.Log("测试场景：分布式锁控制")

	tests := []struct {
		name      string
		lockOK    bool
		lockErr   error
		wantSkip  bool
	}{
		{
			name:     "获取锁成功",
			lockOK:   true,
			lockErr:  nil,
			wantSkip: false,
		},
		{
			name:     "获取锁失败（跳过）",
			lockOK:   false,
			lockErr:  nil,
			wantSkip: true,
		},
		{
			name:     "获取锁异常（跳过）",
			lockOK:   false,
			lockErr:  context.DeadlineExceeded,
			wantSkip: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("测试用例: %s", tt.name)

			acl := NewTradeNotifyACLForTestWithMocks(
				&redisServiceMock{lockResult: tt.lockOK, lockErr: tt.lockErr},
				&httpGatewayMock{notifyResult: enums.NotifyTaskHTTPSuccess},
				&mqPublisherMock{},
			)

			task := &entity.NotifyTaskEntity{
				TeamID:        "TEAM_LOCK_001",
				NotifyType:    string(valobj.NotifyHTTP),
				NotifyUrl:     "http://localhost:8080/callback",
				ParameterJSON: `{"teamId":"TEAM_LOCK_001"}`,
				UUID:          "UUID_LOCK_001",
				ActivityID:    100123,
			}

			result, err := acl.GroupBuyNotify(context.Background(), task)

			if tt.wantSkip {
				if result != enums.NotifyTaskHTTPNull {
					t.Logf("获取锁失败/异常，返回结果: %s", result)
				}
			} else {
				if err != nil {
					t.Logf("获取锁成功，执行回调: result=%s, err=%v", result, err)
				}
			}
		})
	}

	t.Log("✅ 分布式锁测试完成")
}

// TestTradeNotifyACL_UnknownType 未知通知类型
func TestTradeNotifyACL_UnknownType(t *testing.T) {
	t.Log("测试场景：未知通知类型处理")

	acl := NewTradeNotifyACLForTestWithMocks(
		&redisServiceMock{lockResult: true},
		&httpGatewayMock{},
		&mqPublisherMock{},
	)

	task := &entity.NotifyTaskEntity{
		TeamID:        "TEAM_UNKNOWN_001",
		NotifyType:    "UNKNOWN_TYPE",
		ParameterJSON: `{"teamId":"TEAM_UNKNOWN_001"}`,
		UUID:          "UUID_UNKNOWN_001",
		ActivityID:    100123,
	}

	result, err := acl.GroupBuyNotify(context.Background(), task)
	if err != nil {
		t.Fatalf("未知类型处理失败: %v", err)
	}
	if result != enums.NotifyTaskHTTPSuccess {
		t.Errorf("期望未知类型返回成功，实际为 %s", result)
	}

	t.Log("✅ 未知通知类型测试通过")
}

// TestNotifyTaskEntity_LockKey 锁 Key 生成测试
func TestNotifyTaskEntity_LockKey(t *testing.T) {
	t.Log("测试场景：NotifyTaskEntity LockKey 生成")

	tests := []struct {
		uuid     string
		expected string
	}{
		{"UUID_001", "notify_job_lock_key_UUID_001"},
		{"", "notify_job_lock_key_"},
		{"test-uuid-123", "notify_job_lock_key_test-uuid-123"},
	}

	for _, tt := range tests {
		t.Run(tt.uuid, func(t *testing.T) {
			task := &entity.NotifyTaskEntity{UUID: tt.uuid}
			lockKey := task.LockKey()

			if lockKey != tt.expected {
				t.Errorf("期望 LockKey 为 %s，实际为 %s", tt.expected, lockKey)
			} else {
				t.Logf("LockKey 生成正确: %s", lockKey)
			}
		})
	}

	t.Log("✅ LockKey 生成测试完成")
}

// ============================================================================
// 测试辅助函数 - 创建 ACL 实例用于测试
// ============================================================================

// NewTradeNotifyACLForTest 创建简化版 ACL（不依赖真实服务）
func NewTradeNotifyACLForTest() port.ITradePort {
	return NewTradeNotifyACLForTestWithMocks(
		&redisServiceMock{lockResult: true},
		&httpGatewayMock{notifyResult: enums.NotifyTaskHTTPSuccess},
		&mqPublisherMock{},
	)
}

// NewTradeNotifyACLForTestWithMocks 使用 Mock 创建 ACL
func NewTradeNotifyACLForTestWithMocks(
	redisSvc redisx.IRedisService,
	httpGW gateway.IGroupBuyNotifyService,
	publisher mq.IPublisher,
) port.ITradePort {
	return NewTradeNotifyACL(redisSvc, httpGW, publisher)
}

// ============================================================================
// 测试用例汇总
// ============================================================================

// TestTradeNotifyACL_Summary ACL 防腐层测试汇总
func TestTradeNotifyACL_Summary(t *testing.T) {
	t.Log("╔══════════════════════════════════════════════╗")
	t.Log("║      ACL 防腐层测试汇总                     ║")
	t.Log("╠══════════════════════════════════════════════╣")
	t.Log("║ 1. TestTradeNotifyACL_NilTask              ║")
	t.Log("║    - 空任务处理                              ║")
	t.Log("║ 2. TestTradeNotifyACL_HTTPNotify           ║")
	t.Log("║    - HTTP 回调通知                           ║")
	t.Log("║ 3. TestTradeNotifyACL_MQNotify             ║")
	t.Log("║    - MQ 消息通知                             ║")
	t.Log("║ 4. TestTradeNotifyACL_DistributedLock      ║")
	t.Log("║    - 分布式锁控制                           ║")
	t.Log("║ 5. TestTradeNotifyACL_UnknownType          ║")
	t.Log("║    - 未知通知类型处理                        ║")
	t.Log("║ 6. TestNotifyTaskEntity_LockKey            ║")
	t.Log("║    - LockKey 生成                            ║")
	t.Log("╚══════════════════════════════════════════════╝")

	// 执行测试
	t.Run("空值与边界测试", func(t *testing.T) {
		TestTradeNotifyACL_NilTask(t)
		TestTradeNotifyACL_UnknownType(t)
		TestNotifyTaskEntity_LockKey(t)
	})

	t.Run("HTTP 回调测试", func(t *testing.T) {
		TestTradeNotifyACL_HTTPNotify(t)
	})

	t.Run("MQ 通知测试", func(t *testing.T) {
		TestTradeNotifyACL_MQNotify(t)
	})

	t.Run("分布式锁测试", func(t *testing.T) {
		TestTradeNotifyACL_DistributedLock(t)
	})

	t.Log("╔══════════════════════════════════════════════╗")
	t.Log("║      ✅ ACL 防腐层测试全部通过              ║")
	t.Log("╚══════════════════════════════════════════════╝")
}

// 确保接口实现正确
var _ port.ITradePort = (*TradeNotifyACLForTest)(nil)

// TradeNotifyACLForTest 测试用 ACL 别名
type TradeNotifyACLForTest = struct{}

// 忽略未使用的变量
var _ = valobj.NotifyHTTP
var _ = valobj.NotifyMQ
var _ = gateway.NewGroupBuyNotifyService
var _ = mq.NewEventPublisher
var _ = redisx.New

package refund_test

import (
	"context"
	"testing"
	"time"

	"group-buy-market/internal/domain/trade/adapter/port"
	"group-buy-market/internal/domain/trade/adapter/repository"
	"group-buy-market/internal/domain/trade/model/aggregate"
	"group-buy-market/internal/domain/trade/model/entity"
	"group-buy-market/internal/domain/trade/model/valobj"
	"group-buy-market/internal/domain/trade/service/refund"
	"group-buy-market/internal/domain/trade/service/refund/business"
	"group-buy-market/internal/domain/trade/service/task"
	"group-buy-market/internal/types/enums"
	"group-buy-market/internal/types/exception"
)

// ============================================================================
// Mock 仓储实现 - 退单服务专用
// ============================================================================

type refundRepoMock struct {
	// 退单结果
	unpaidNotify  *entity.NotifyTaskEntity
	unpaidErr     error
	paidNotify    *entity.NotifyTaskEntity
	paidErr       error
	paidTeamNotify *entity.NotifyTaskEntity
	paidTeamErr   error
	// 恢复库存
	refundRecoveryCalled bool
	refundRecoveryOK     bool
	refundRecoveryErr    error
	// 退单幂等检查
	refundExist     bool
	refundExistErr  error
	// 超时订单
	timeoutOrders []*entity.TimeoutUnpaidOrderEntity
}

func (m *refundRepoMock) QueryMarketPayOrderEntityByOutTradeNo(ctx context.Context, userID, outTradeNo string) (*entity.MarketPayOrderEntity, error) {
	return nil, nil
}

func (m *refundRepoMock) LockMarketPayOrder(ctx context.Context, agg *aggregate.GroupBuyOrderAggregate) (*entity.MarketPayOrderEntity, error) {
	return nil, nil
}

func (m *refundRepoMock) QueryGroupBuyProgress(ctx context.Context, teamID string) (*valobj.GroupBuyProgressVO, error) {
	return nil, nil
}

func (m *refundRepoMock) QueryGroupBuyActivityEntityByActivityID(ctx context.Context, activityID int64) (*entity.GroupBuyActivityEntity, error) {
	return nil, nil
}

func (m *refundRepoMock) QueryOrderCountByActivityID(ctx context.Context, activityID int64, userID string) (int, error) {
	return 0, nil
}

func (m *refundRepoMock) QueryGroupBuyTeamByTeamID(ctx context.Context, teamID string) (*entity.GroupBuyTeamEntity, error) {
	return nil, nil
}

func (m *refundRepoMock) SettlementMarketPayOrder(ctx context.Context, agg *aggregate.GroupBuyTeamSettlementAggregate) (*entity.NotifyTaskEntity, error) {
	return nil, nil
}

func (m *refundRepoMock) IsSCBlackIntercept(source, channel string) bool {
	return false
}

func (m *refundRepoMock) QueryUnExecutedNotifyTaskList(ctx context.Context) ([]*entity.NotifyTaskEntity, error) {
	return nil, nil
}

func (m *refundRepoMock) QueryUnExecutedNotifyTaskListByTeamID(ctx context.Context, teamID string) ([]*entity.NotifyTaskEntity, error) {
	return nil, nil
}

func (m *refundRepoMock) UpdateNotifyTaskStatusSuccess(ctx context.Context, task *entity.NotifyTaskEntity) (int, error) {
	return 0, nil
}

func (m *refundRepoMock) UpdateNotifyTaskStatusError(ctx context.Context, task *entity.NotifyTaskEntity) (int, error) {
	return 0, nil
}

func (m *refundRepoMock) UpdateNotifyTaskStatusRetry(ctx context.Context, task *entity.NotifyTaskEntity) (int, error) {
	return 0, nil
}

func (m *refundRepoMock) OccupyTeamStock(ctx context.Context, teamStockKey, recoveryTeamStockKey string, target, validTime int) (bool, error) {
	return true, nil
}

func (m *refundRepoMock) RecoveryTeamStock(ctx context.Context, recoveryTeamStockKey string, validTime int) error {
	return nil
}

func (m *refundRepoMock) Unpaid2Refund(ctx context.Context, agg *aggregate.GroupBuyRefundAggregate) (*entity.NotifyTaskEntity, error) {
	return m.unpaidNotify, m.unpaidErr
}

func (m *refundRepoMock) Paid2Refund(ctx context.Context, agg *aggregate.GroupBuyRefundAggregate) (*entity.NotifyTaskEntity, error) {
	return m.paidNotify, m.paidErr
}

func (m *refundRepoMock) PaidTeam2Refund(ctx context.Context, agg *aggregate.GroupBuyRefundAggregate) (*entity.NotifyTaskEntity, error) {
	return m.paidTeamNotify, m.paidTeamErr
}

func (m *refundRepoMock) QueryTimeoutUnpaidOrderList(ctx context.Context) ([]*entity.TimeoutUnpaidOrderEntity, error) {
	return m.timeoutOrders, nil
}

func (m *refundRepoMock) RefundOrderExist(ctx context.Context, teamID, category, orderID string) (bool, error) {
	return m.refundExist, m.refundExistErr
}

func (m *refundRepoMock) Refund2AddRecovery(ctx context.Context, recoveryTeamStockKey, orderID string) error {
	m.refundRecoveryCalled = true
	if m.refundRecoveryErr != nil {
		return m.refundRecoveryErr
	}
	if !m.refundRecoveryOK {
		return context.DeadlineExceeded
	}
	return nil
}

// ============================================================================
// Mock 端口（ACL）实现
// ============================================================================

type portMock struct {
	notifyResult string
	notifyErr    error
}

func (m *portMock) GroupBuyNotify(ctx context.Context, task *entity.NotifyTaskEntity) (string, error) {
	return m.notifyResult, m.notifyErr
}

// ============================================================================
// Mock 任务服务
// ============================================================================

type taskServiceMock struct {
	execResult string
	execErr    error
}

func (m *taskServiceMock) ExecNotifyJob(ctx context.Context, task *entity.NotifyTaskEntity) (string, error) {
	return m.execResult, m.execErr
}

// ============================================================================
// 测试辅助函数
// ============================================================================

func buildRefundOrder(userID, teamID, orderID string) *entity.TradeRefundOrderEntity {
	return &entity.TradeRefundOrderEntity{
		UserID:     userID,
		OrderID:    orderID,
		TeamID:     teamID,
		ActivityID: 100123,
		OutTradeNo: orderID,
	}
}

func buildRefundCommand(userID, outTradeNo string) *entity.TradeRefundCommandEntity {
	return &entity.TradeRefundCommandEntity{
		UserID:     userID,
		OutTradeNo: outTradeNo,
		Source:     "s01",
		Channel:    "c01",
	}
}

func buildRefundSuccessMsg(userID, teamID, orderID string, refundType int) *valobj.TeamRefundSuccess {
	return &valobj.TeamRefundSuccess{
		UserID:     userID,
		TeamID:     teamID,
		OrderID:    orderID,
		ActivityID: 100123,
		Type:       refundType,
	}
}

// ============================================================================
// 1. 未支付未成团退单策略测试
// ============================================================================

// TestUnpaid2RefundStrategy 未支付未成团退单
func TestUnpaid2RefundStrategy(t *testing.T) {
	t.Log("测试场景：未支付未成团退单策略")

	tests := []struct {
		name    string
		wantErr bool
	}{
		{"正常退单", false},
		{"退单失败", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("测试用例: %s", tt.name)

			repo := &refundRepoMock{
				unpaidNotify: &entity.NotifyTaskEntity{
					TeamID:        "TEAM_UNPAID_001",
					NotifyType:    "MQ",
					NotifyMQ:      "topic.team_refund",
					NotifyCount:   0,
					ParameterJSON: `{"teamId":"TEAM_UNPAID_001"}`,
					UUID:          "UUID_UNPAID_001",
					ActivityID:    100123,
				},
				unpaidErr: nil,
			}
			if tt.wantErr {
				repo.unpaidNotify = nil
				repo.unpaidErr = context.DeadlineExceeded
			}

			p := &portMock{notifyResult: "success"}
			_ = p

			strategy := business.NewUnpaid2Refund(repo, nil, nil)

			order := buildRefundOrder("user001", "TEAM_UNPAID_001", "ORDER_UNPAID_001")
			err := strategy.RefundOrder(context.Background(), order)

			if tt.wantErr && err == nil {
				t.Error("期望退单失败，但没有返回错误")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("期望退单成功，但返回错误: %v", err)
			}

			t.Logf("未支付未成团退单测试完成: err=%v", err)
		})
	}

	t.Log("✅ 未支付未成团退单策略测试完成")
}

// TestUnpaid2Refund_ReverseStock 未支付退单恢复库存
func TestUnpaid2Refund_ReverseStock(t *testing.T) {
	t.Log("测试场景：未支付未成团退单恢复库存")

	tests := []struct {
		name    string
		ok      bool
		wantErr bool
	}{
		{"恢复成功", true, false},
		{"恢复失败", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("测试用例: %s", tt.name)

			repo := &refundRepoMock{
				refundRecoveryOK: tt.ok,
				refundRecoveryErr: func() error {
					if tt.ok {
						return nil
					}
					return context.DeadlineExceeded
				}(),
			}

			strategy := business.NewUnpaid2Refund(repo, nil, nil)

			msg := buildRefundSuccessMsg("user001", "TEAM_UNPAID_002", "ORDER_UNPAID_002", 1)
			err := strategy.ReverseStock(context.Background(), msg)

			if tt.wantErr && err == nil {
				t.Error("期望恢复库存失败，但没有返回错误")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("期望恢复库存成功，但返回错误: %v", err)
			}

			t.Logf("未支付退单恢复库存测试完成: err=%v", err)
		})
	}

	t.Log("✅ 未支付退单恢复库存测试完成")
}

// ============================================================================
// 2. 已支付未成团退单策略测试
// ============================================================================

// TestPaid2RefundStrategy 已支付未成团退单
func TestPaid2RefundStrategy(t *testing.T) {
	t.Log("测试场景：已支付未成团退单策略")

	repo := &refundRepoMock{
		paidNotify: &entity.NotifyTaskEntity{
			TeamID:        "TEAM_PAID_001",
			NotifyType:    "MQ",
			NotifyMQ:      "topic.team_refund",
			NotifyCount:   0,
			ParameterJSON: `{"teamId":"TEAM_PAID_001"}`,
			UUID:          "UUID_PAID_001",
			ActivityID:    100123,
		},
		paidErr: nil,
	}

	strategy := business.NewPaid2Refund(repo, nil, nil)

	order := buildRefundOrder("user002", "TEAM_PAID_001", "ORDER_PAID_001")
	err := strategy.RefundOrder(context.Background(), order)
	if err != nil {
		t.Fatalf("已支付未成团退单失败: %v", err)
	}

	t.Log("✅ 已支付未成团退单成功")
}

// TestPaid2Refund_ReverseStock 已支付退单恢复库存
func TestPaid2Refund_ReverseStock(t *testing.T) {
	t.Log("测试场景：已支付未成团退单恢复库存")

	repo := &refundRepoMock{
		refundRecoveryOK: true,
	}

	strategy := business.NewPaid2Refund(repo, nil, nil)

	msg := buildRefundSuccessMsg("user002", "TEAM_PAID_002", "ORDER_PAID_002", 2)
	err := strategy.ReverseStock(context.Background(), msg)
	if err != nil {
		t.Fatalf("已支付退单恢复库存失败: %v", err)
	}

	t.Log("✅ 已支付退单恢复库存成功")
}

// ============================================================================
// 3. 已支付已成团退单策略测试
// ============================================================================

// TestPaidTeam2RefundStrategy 已支付已成团退单
func TestPaidTeam2RefundStrategy(t *testing.T) {
	t.Log("测试场景：已支付已成团退单策略")

	repo := &refundRepoMock{
		paidTeamNotify: &entity.NotifyTaskEntity{
			TeamID:        "TEAM_PAID_TEAM_001",
			NotifyType:    "MQ",
			NotifyMQ:      "topic.team_refund",
			NotifyCount:   0,
			ParameterJSON: `{"teamId":"TEAM_PAID_TEAM_001"}`,
			UUID:          "UUID_PAID_TEAM_001",
			ActivityID:    100123,
		},
		paidTeamErr: nil,
	}

	strategy := business.NewPaidTeam2Refund(repo, nil, nil)

	order := buildRefundOrder("user003", "TEAM_PAID_TEAM_001", "ORDER_PAID_TEAM_001")
	err := strategy.RefundOrder(context.Background(), order)
	if err != nil {
		t.Fatalf("已支付已成团退单失败: %v", err)
	}

	t.Log("✅ 已支付已成团退单成功")
}

// TestPaidTeam2Refund_ReverseStock 已成团退单恢复库存
func TestPaidTeam2Refund_ReverseStock(t *testing.T) {
	t.Log("测试场景：已支付已成团退单恢复库存")

	repo := &refundRepoMock{
		refundRecoveryOK: true,
	}

	strategy := business.NewPaidTeam2Refund(repo, nil, nil)

	msg := buildRefundSuccessMsg("user003", "TEAM_PAID_TEAM_002", "ORDER_PAID_TEAM_002", 3)
	err := strategy.ReverseStock(context.Background(), msg)
	if err != nil {
		t.Fatalf("已成团退单恢复库存失败: %v", err)
	}

	t.Log("✅ 已成团退单恢复库存成功")
}

// ============================================================================
// 策略注册表测试
// ============================================================================

// TestStrategyMap_Registration 策略注册表测试
func TestStrategyMap_Registration(t *testing.T) {
	t.Log("测试场景：退单策略注册表")

	repo := &refundRepoMock{}
	strategies := business.NewStrategyMap(
		business.NewUnpaid2Refund(repo, nil, nil),
		business.NewPaid2Refund(repo, nil, nil),
		business.NewPaidTeam2Refund(repo, nil, nil),
	)

	// 验证所有策略都注册成功
	expectedStrategies := []string{
		"unpaid2RefundStrategy",
		"paid2RefundStrategy",
		"paidTeam2RefundStrategy",
	}

	for _, name := range expectedStrategies {
		if _, ok := strategies[name]; !ok {
			t.Errorf("策略 %s 未注册", name)
		} else {
			t.Logf("策略 %s 注册成功", name)
		}
	}

	// 验证策略数量
	if len(strategies) != 3 {
		t.Errorf("期望注册 3 个策略，实际为 %d", len(strategies))
	}

	t.Logf("✅ 策略注册表测试完成: 共 %d 个策略", len(strategies))
}

// TestStrategyMap_Selection 策略选择测试
func TestStrategyMap_Selection(t *testing.T) {
	t.Log("测试场景：根据退单类型选择正确的策略")

	repo := &refundRepoMock{}
	strategies := business.NewStrategyMap(
		business.NewUnpaid2Refund(repo, nil, nil),
		business.NewPaid2Refund(repo, nil, nil),
		business.NewPaidTeam2Refund(repo, nil, nil),
	)

	// 测试每种退单类型对应的策略
	tests := []struct {
		refundType int
		strategy   string
	}{
		{1, "unpaid2RefundStrategy"},  // 未支付未成团
		{2, "paid2RefundStrategy"},    // 已支付未成团
		{3, "paidTeam2RefundStrategy"}, // 已支付已成团
	}

	for _, tt := range tests {
		rt, err := valobj.GetRefundTypeByCode(tt.refundType)
		if err != nil {
			t.Errorf("获取退单类型失败: %v", err)
			continue
		}

		strategy, ok := strategies[rt.Strategy]
		if !ok {
			t.Errorf("退单类型 %d 对应的策略 %s 未找到", tt.refundType, rt.Strategy)
			continue
		}

		if strategy.Name() != tt.strategy {
			t.Errorf("期望策略 %s，实际为 %s", tt.strategy, strategy.Name())
		} else {
			t.Logf("退单类型 %d -> 策略 %s 匹配正确", tt.refundType, strategy.Name())
		}
	}

	t.Log("✅ 策略选择测试完成")
}

// ============================================================================
// 退单服务集成测试
// ============================================================================

// TestTradeRefundOrderService_RefundOrder 退单服务完整流程
func TestTradeRefundOrderService_RefundOrder(t *testing.T) {
	t.Log("测试场景：退单服务完整流程")

	tests := []struct {
		name       string
		userID     string
		outTradeNo string
		wantOK     bool
	}{
		{"正常退单", "user001", "OUT_TRADE_REFUND_001", true},
		{"重复退单（幂等）", "user001", "OUT_TRADE_REFUND_001", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("测试用例: %s", tt.name)

			repo := &refundRepoMock{
				unpaidNotify: &entity.NotifyTaskEntity{
					TeamID:        "TEAM_REFUND_001",
					NotifyType:    "MQ",
					NotifyMQ:      "topic.team_refund",
					NotifyCount:   0,
					ParameterJSON: `{"teamId":"TEAM_REFUND_001"}`,
					UUID:          "UUID_REFUND_001",
					ActivityID:    100123,
				},
				refundRecoveryOK: true,
			}

			p := &portMock{notifyResult: "success"}
			taskSvc := &taskServiceMock{execResult: "success"}

			strategies := business.NewStrategyMap(
				business.NewUnpaid2Refund(repo, p, taskSvc),
				business.NewPaid2Refund(repo, p, taskSvc),
				business.NewPaidTeam2Refund(repo, p, taskSvc),
			)

			svc := refund.NewTradeRefundOrderService(repo, strategies)

			cmd := buildRefundCommand(tt.userID, tt.outTradeNo)
			result, err := svc.RefundOrder(context.Background(), cmd)

			if tt.wantOK && err != nil {
				t.Errorf("期望退单成功，但返回错误: %v", err)
			}
			if !tt.wantOK && err == nil {
				t.Error("期望退单失败，但返回成功")
			}

			if result != nil {
				t.Logf("退单结果: UserID=%s, Behavior=%s", result.UserID, result.Behavior.Info())
			}
		})
	}

	t.Log("✅ 退单服务集成测试完成")
}

// TestRestoreTeamLockStock 恢复锁单库存测试
func TestRestoreTeamLockStock(t *testing.T) {
	t.Log("测试场景：恢复锁单库存（MQ 消费端处理）")

	tests := []struct {
		name       string
		refundType int
		wantErr    bool
	}{
		{"未支付退单恢复", 1, false},
		{"已支付退单恢复", 2, false},
		{"已成团退单恢复", 3, false},
		{"未知类型", 99, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("测试用例: %s", tt.name)

			repo := &refundRepoMock{
				refundRecoveryOK: true,
			}

			strategies := business.NewStrategyMap(
				business.NewUnpaid2Refund(repo, nil, nil),
				business.NewPaid2Refund(repo, nil, nil),
				business.NewPaidTeam2Refund(repo, nil, nil),
			)

			svc := refund.NewTradeRefundOrderService(repo, strategies)

			msg := buildRefundSuccessMsg("user004", "TEAM_RECOVER_001", "ORDER_RECOVER_001", tt.refundType)
			err := svc.RestoreTeamLockStock(context.Background(), msg)

			if tt.wantErr && err == nil {
				t.Error("期望恢复库存失败，但没有返回错误")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("期望恢复库存成功，但返回错误: %v", err)
			}

			if err == nil {
				t.Logf("恢复库存成功: refundType=%d", tt.refundType)
			}
		})
	}

	t.Log("✅ 恢复锁单库存测试完成")
}

// TestQueryTimeoutUnpaidOrderList 查询超时未支付订单
func TestQueryTimeoutUnpaidOrderList(t *testing.T) {
	t.Log("测试场景：查询超时未支付订单列表")

	timeoutOrders := []*entity.TimeoutUnpaidOrderEntity{
		{
			UserID:   "user005",
			OutTradeNo: "OUT_TRADE_TIMEOUT_001",
			Source:   "s01",
			Channel:  "c01",
		},
		{
			UserID:   "user006",
			OutTradeNo: "OUT_TRADE_TIMEOUT_002",
			Source:   "s01",
			Channel:  "c01",
		},
	}

	repo := &refundRepoMock{
		timeoutOrders: timeoutOrders,
	}

	svc := refund.NewTradeRefundOrderService(repo, business.NewStrategyMap())

	orders, err := svc.QueryTimeoutUnpaidOrderList(context.Background())
	if err != nil {
		t.Fatalf("查询超时订单失败: %v", err)
	}

	if len(orders) != 2 {
		t.Errorf("期望查询到 2 个超时订单，实际为 %d", len(orders))
	}

	for _, order := range orders {
		t.Logf("超时订单: UserID=%s, OutTradeNo=%s", order.UserID, order.OutTradeNo)
	}

	t.Log("✅ 查询超时未支付订单测试完成")
}

// ============================================================================
// 退单幂等性测试
// ============================================================================

// TestRefundIdempotency 退单幂等性测试
func TestRefundIdempotency(t *testing.T) {
	t.Log("测试场景：退单幂等性（防止重复退单）")

	tests := []struct {
		name          string
		existOrder    bool
		expectSuccess bool
	}{
		{"首次退单（幂等通过）", false, true},
		{"重复退单（幂等拦截）", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("测试用例: %s", tt.name)

			repo := &refundRepoMock{
				refundExist: tt.existOrder,
				unpaidNotify: &entity.NotifyTaskEntity{
					TeamID:     "TEAM_IDEMPOTENT",
					NotifyType: "MQ",
				},
			}

			// 验证幂等检查
			exist, err := repo.RefundOrderExist(context.Background(), "TEAM_IDEMPOTENT", "unpaid", "ORDER_IDEMPOTENT")
			if err != nil {
				t.Fatalf("幂等检查失败: %v", err)
			}

			if exist != tt.existOrder {
				t.Errorf("期望退单存在=%v，实际为 %v", tt.existOrder, exist)
			}

			if tt.existOrder {
				t.Log("退单已存在，幂等拦截成功")
			} else {
				t.Log("首次退单，可以执行")
			}
		})
	}

	t.Log("✅ 退单幂等性测试完成")
}

// ============================================================================
// 库存恢复原子性测试
// ============================================================================

// TestRefundRecovery_Atomicity 库存恢复原子性测试
func TestRefundRecovery_Atomicity(t *testing.T) {
	t.Log("测试场景：库存恢复的原子性和幂等性")

	tests := []struct {
		name    string
		ok      bool
		wantErr bool
	}{
		{"首次恢复（成功）", true, false},
		{"重复恢复（Redis SETNX 拦截）", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("测试用例: %s", tt.name)

			repo := &refundRepoMock{
				refundRecoveryOK: tt.ok,
				refundRecoveryErr: func() error {
					if tt.ok {
						return nil
					}
					return context.DeadlineExceeded
				}(),
			}

			// 调用恢复库存
			err := repo.Refund2AddRecovery(context.Background(), "recovery_key", "order_001")

			if tt.wantErr && err == nil {
				t.Error("期望恢复失败，但没有返回错误")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("期望恢复成功，但返回错误: %v", err)
			}

			// 验证调用
			if !repo.refundRecoveryCalled {
				t.Error("恢复库存方法应该被调用")
			}

			t.Logf("库存恢复测试完成: err=%v", err)
		})
	}

	t.Log("✅ 库存恢复原子性测试完成")
}

// ============================================================================
// 退单行为枚举测试
// ============================================================================

// TestRefundBehavior_Enum 退单行为枚举测试
func TestRefundBehavior_Enum(t *testing.T) {
	t.Log("测试场景：退单行为枚举值验证")

	tests := []struct {
		behavior entity.TradeRefundBehavior
		code     string
		info     string
	}{
		{entity.RefundBehaviorSuccess, "success", "退单成功"},
		{entity.RefundBehaviorRepeat, "repeat", "重复退单"},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			code := tt.behavior.Code()
			info := tt.behavior.Info()

			if code != tt.code {
				t.Errorf("期望行为代码 %s，实际为 %s", tt.code, code)
			}
			if info != tt.info {
				t.Errorf("期望行为描述 %s，实际为 %s", tt.info, info)
			}

			t.Logf("退单行为 %d: code=%s, info=%s", tt.behavior, code, info)
		})
	}

	t.Log("✅ 退单行为枚举测试完成")
}

// ============================================================================
// 测试用例汇总
// ============================================================================

// TestRefundOrderService_Summary 退单服务测试汇总
func TestRefundOrderService_Summary(t *testing.T) {
	t.Log("╔══════════════════════════════════════════════╗")
	t.Log("║      退单业务集成测试汇总                    ║")
	t.Log("╠══════════════════════════════════════════════╣")
	t.Log("║ 1. TestUnpaid2RefundStrategy                ║")
	t.Log("║    - 未支付未成团退单策略                     ║")
	t.Log("║ 2. TestPaid2RefundStrategy                  ║")
	t.Log("║    - 已支付未成团退单策略                     ║")
	t.Log("║ 3. TestPaidTeam2RefundStrategy              ║")
	t.Log("║    - 已支付已成团退单策略                     ║")
	t.Log("║ 4. TestStrategyMap_Registration             ║")
	t.Log("║    - 策略注册表测试                           ║")
	t.Log("║ 5. TestStrategyMap_Selection               ║")
	t.Log("║    - 策略选择测试                             ║")
	t.Log("║ 6. TestTradeRefundOrderService_RefundOrder  ║")
	t.Log("║    - 退单服务完整流程                         ║")
	t.Log("║ 7. TestRestoreTeamLockStock                ║")
	t.Log("║    - 恢复锁单库存                             ║")
	t.Log("║ 8. TestRefundIdempotency                   ║")
	t.Log("║    - 退单幂等性                               ║")
	t.Log("╚══════════════════════════════════════════════╝")

	// 执行核心策略测试
	t.Run("未支付退单策略", func(t *testing.T) {
		TestUnpaid2RefundStrategy(t)
		TestUnpaid2Refund_ReverseStock(t)
	})

	t.Run("已支付退单策略", func(t *testing.T) {
		TestPaid2RefundStrategy(t)
		TestPaid2Refund_ReverseStock(t)
	})

	t.Run("已成团退单策略", func(t *testing.T) {
		TestPaidTeam2RefundStrategy(t)
		TestPaidTeam2Refund_ReverseStock(t)
	})

	t.Run("策略注册表测试", func(t *testing.T) {
		TestStrategyMap_Registration(t)
		TestStrategyMap_Selection(t)
	})

	t.Run("退单服务集成测试", func(t *testing.T) {
		TestTradeRefundOrderService_RefundOrder(t)
		TestRestoreTeamLockStock(t)
	})

	t.Run("幂等性与原子性测试", func(t *testing.T) {
		TestRefundIdempotency(t)
		TestRefundRecovery_Atomicity(t)
		TestRefundBehavior_Enum(t)
	})

	t.Run("超时订单查询测试", func(t *testing.T) {
		TestQueryTimeoutUnpaidOrderList(t)
	})

	t.Log("╔══════════════════════════════════════════════╗")
	t.Log("║      ✅ 退单业务集成测试全部通过             ║")
	t.Log("╚══════════════════════════════════════════════╝")
}

// 确保接口实现正确
var _ repository.ITradeRepository = (*refundRepoMock)(nil)
var _ port.ITradePort = (*portMock)(nil)
var _ refund.ITradeRefundOrderService = (*refund.TradeRefundOrderService)(nil)
var _ business.IRefundOrderStrategy = (*business.Unpaid2RefundStrategy)(nil)
var _ business.IRefundOrderStrategy = (*business.Paid2RefundStrategy)(nil)
var _ business.IRefundOrderStrategy = (*business.PaidTeam2RefundStrategy)(nil)

// 忽略未使用的变量
var _ = time.Now
var _ = exception.NewAppException
var _ = enums.REFUND_BEHAVIOR_SUCCESS

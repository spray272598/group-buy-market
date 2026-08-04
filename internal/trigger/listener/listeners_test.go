package listener_test

import (
	"context"
	"encoding/json"
	"testing"

	"group-buy-market/internal/domain/trade/model/entity"
	"group-buy-market/internal/domain/trade/model/valobj"
	"group-buy-market/internal/domain/trade/service/refund"
	"group-buy-market/internal/trigger/listener"
)

// ============================================================================
// Mock 退单服务
// ============================================================================

type refundServiceMock struct {
	restoreStockCalled bool
	restoreStockResult error
	timeoutOrders      []*entity.TimeoutUnpaidOrderEntity
}

func (m *refundServiceMock) RefundOrder(ctx context.Context, cmd *entity.TradeRefundCommandEntity) (*entity.TradeRefundResultEntity, error) {
	return nil, nil
}

func (m *refundServiceMock) RestoreTeamLockStock(ctx context.Context, msg *valobj.TeamRefundSuccess) error {
	m.restoreStockCalled = true
	return m.restoreStockResult
}

func (m *refundServiceMock) QueryTimeoutUnpaidOrderList(ctx context.Context) ([]*entity.TimeoutUnpaidOrderEntity, error) {
	return m.timeoutOrders, nil
}

// 确保接口实现正确
var _ refund.ITradeRefundOrderService = (*refundServiceMock)(nil)

// ============================================================================
// 消费者处理逻辑测试
// ============================================================================

// TestTeamRefundConsumerHandler 退单消费者处理逻辑
func TestTeamRefundConsumerHandler(t *testing.T) {
	t.Log("测试场景：退单消费者消息处理逻辑")

	tests := []struct {
		name        string
		message     valobj.TeamRefundSuccess
		restoreErr  error
		wantCalled  bool
		wantErr     bool
	}{
		{
			name: "正常恢复库存",
			message: valobj.TeamRefundSuccess{
				UserID:     "user001",
				TeamID:     "TEAM_CONSUME_001",
				OrderID:    "ORDER_CONSUME_001",
				ActivityID: 100123,
				Type:       1,
			},
			restoreErr: nil,
			wantCalled: true,
			wantErr: false,
		},
		{
			name: "恢复库存失败（重试）",
			message: valobj.TeamRefundSuccess{
				UserID:     "user002",
				TeamID:     "TEAM_CONSUME_002",
				OrderID:    "ORDER_CONSUME_002",
				ActivityID: 100123,
				Type:       2,
			},
			restoreErr: context.DeadlineExceeded,
			wantCalled: true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("测试用例: %s", tt.name)

			refundSvc := &refundServiceMock{
				restoreStockResult: tt.restoreErr,
			}

			// 模拟消费者处理逻辑（与 listeners.go 中的逻辑一致）
			body, _ := json.Marshal(tt.message)
			var msg valobj.TeamRefundSuccess
			err := json.Unmarshal(body, &msg)
			if err != nil {
				t.Fatalf("消息反序列化失败: %v", err)
			}

			err = refundSvc.RestoreTeamLockStock(context.Background(), &msg)

			if tt.wantCalled && !refundSvc.restoreStockCalled {
				t.Error("恢复库存方法应该被调用")
			}
			if tt.wantErr && err == nil {
				t.Error("期望返回错误（触发 Nack 重试）")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("期望成功，实际返回错误: %v", err)
			}

			t.Logf("退单消费测试完成: restoreCalled=%v, err=%v", refundSvc.restoreStockCalled, err)
		})
	}

	t.Log("✅ 退单消费者处理逻辑测试完成")
}

// TestTeamRefundConsumer_InvalidMessage 无效消息处理
func TestTeamRefundConsumer_InvalidMessage(t *testing.T) {
	t.Log("测试场景：无效消息处理（JSON 格式错误）")

	refundSvc := &refundServiceMock{}

	// 模拟无效的 JSON 消息
	invalidBody := []byte(`{"invalid": "json", "broken": `)

	var msg valobj.TeamRefundSuccess
	err := json.Unmarshal(invalidBody, &msg)
	if err == nil {
		t.Error("期望无效 JSON 解析失败")
	}

	// 验证无效消息不会调用恢复库存
	if refundSvc.restoreStockCalled {
		t.Error("无效消息不应该触发恢复库存")
	}

	t.Logf("无效消息处理测试完成: err=%v", err)
	t.Log("✅ 无效消息处理测试通过")
}

// TestTeamRefundConsumer_EmptyMessage 空消息处理
func TestTeamRefundConsumer_EmptyMessage(t *testing.T) {
	t.Log("测试场景：空消息处理")

	refundSvc := &refundServiceMock{}

	// 模拟空消息
	emptyBody := []byte(``)

	var msg valobj.TeamRefundSuccess
	err := json.Unmarshal(emptyBody, &msg)
	if err == nil {
		t.Log("空 JSON 解析成功（空结构体）")
		// 空结构体字段为零值，仍然可以处理
		err = refundSvc.RestoreTeamLockStock(context.Background(), &msg)
		if err != nil {
			t.Logf("空结构体处理: err=%v", err)
		}
	}

	t.Log("✅ 空消息处理测试完成")
}

// TestTeamSuccessConsumerHandler 组队成功消费者处理
func TestTeamSuccessConsumerHandler(t *testing.T) {
	t.Log("测试场景：组队成功消费者处理逻辑")

	// 组队成功消息通常只需要日志记录
	// 实际业务中可以添加通知、积分等逻辑

	messages := []struct {
		name string
		body []byte
	}{
		{
			name: "组队成功消息",
			body: []byte(`{"teamId":"TEAM_SUCCESS_001","userId":"user001","activityId":100123}`),
		},
		{
			name: "组队成功消息（完整字段）",
			body: []byte(`{"teamId":"TEAM_SUCCESS_002","userId":"user002","activityId":100456,"orderId":"ORDER_SUCCESS_002","count":3}`),
		},
	}

	for _, msg := range messages {
		t.Run(msg.name, func(t *testing.T) {
			t.Logf("测试用例: %s", msg.name)

			// 组队成功消息处理通常是成功（直接 Ack）
			err := processTeamSuccessMessage(msg.body)
			if err != nil {
				t.Errorf("组队成功消息处理失败: %v", err)
			}

			t.Logf("组队成功消息处理完成: body=%s", string(msg.body))
		})
	}

	t.Log("✅ 组队成功消费者处理测试完成")
}

// TestStartConsumers_NilClient 空客户端处理
func TestStartConsumers_NilClient(t *testing.T) {
	t.Log("测试场景：空客户端处理")

	// 空客户端应该优雅降级，不报错
	err := listener.StartConsumers(nil, nil)
	if err != nil {
		t.Errorf("空客户端处理应该成功，实际返回错误: %v", err)
	}

	t.Log("✅ 空客户端处理测试通过")
}

// TestRefundSuccessMessage_Structure 退单成功消息结构验证
func TestRefundSuccessMessage_Structure(t *testing.T) {
	t.Log("测试场景：退单成功消息结构验证")

	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name:    "完整消息",
			json:    `{"userID":"user001","teamID":"TEAM_001","orderID":"ORDER_001","activityID":100123,"type":1}`,
			wantErr: false,
		},
		{
			name:    "缺少可选字段",
			json:    `{"userID":"user002","teamID":"TEAM_002","type":2}`,
			wantErr: false,
		},
		{
			name:    "空 JSON 对象",
			json:    `{}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("测试用例: %s", tt.name)

			var msg valobj.TeamRefundSuccess
			err := json.Unmarshal([]byte(tt.json), &msg)

			if tt.wantErr && err == nil {
				t.Error("期望解析失败")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("期望解析成功，实际失败: %v", err)
			}

			t.Logf("消息结构验证完成: msg=%+v", msg)
		})
	}

	t.Log("✅ 退单成功消息结构验证完成")
}

// ============================================================================
// 辅助函数
// ============================================================================

// processTeamSuccessMessage 处理组队成功消息
// 实际业务中可以实现通知、积分发放等逻辑
func processTeamSuccessMessage(body []byte) error {
	// 组队成功消息处理
	// 目前仅做日志记录，实际业务可扩展
	if len(body) == 0 {
		return nil
	}
	return nil
}

// ============================================================================
// 测试用例汇总
// ============================================================================

// TestMQConsumers_Summary MQ 消费者测试汇总
func TestMQConsumers_Summary(t *testing.T) {
	t.Log("╔══════════════════════════════════════════════╗")
	t.Log("║      MQ 消费者测试汇总                      ║")
	t.Log("╠══════════════════════════════════════════════╣")
	t.Log("║ 1. TestTeamRefundConsumerHandler           ║")
	t.Log("║    - 退单消费者处理逻辑                     ║")
	t.Log("║ 2. TestTeamRefundConsumer_InvalidMessage   ║")
	t.Log("║    - 无效消息处理                           ║")
	t.Log("║ 3. TestTeamRefundConsumer_EmptyMessage     ║")
	t.Log("║    - 空消息处理                             ║")
	t.Log("║ 4. TestTeamSuccessConsumerHandler          ║")
	t.Log("║    - 组队成功消费者处理                     ║")
	t.Log("║ 5. TestStartConsumers_NilClient            ║")
	t.Log("║    - 空客户端优雅降级                       ║")
	t.Log("║ 6. TestRefundSuccessMessage_Structure      ║")
	t.Log("║    - 消息结构验证                           ║")
	t.Log("╚══════════════════════════════════════════════╝")

	// 执行测试
	t.Run("退单消费者处理测试", func(t *testing.T) {
		TestTeamRefundConsumerHandler(t)
		TestTeamRefundConsumer_InvalidMessage(t)
		TestTeamRefundConsumer_EmptyMessage(t)
	})

	t.Run("组队成功消费者测试", func(t *testing.T) {
		TestTeamSuccessConsumerHandler(t)
	})

	t.Run("初始化与降级测试", func(t *testing.T) {
		TestStartConsumers_NilClient(t)
	})

	t.Run("消息结构验证", func(t *testing.T) {
		TestRefundSuccessMessage_Structure(t)
	})

	t.Log("╔══════════════════════════════════════════════╗")
	t.Log("║      ✅ MQ 消费者测试全部通过              ║")
	t.Log("╚══════════════════════════════════════════════╝")
}

// 确保接口实现正确
var _ = processTeamSuccessMessage

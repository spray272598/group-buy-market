package filter

import (
	"context"
	"testing"
	"time"

	"group-buy-market/internal/domain/trade/model/aggregate"
	"group-buy-market/internal/domain/trade/model/entity"
	"group-buy-market/internal/domain/trade/model/valobj"
	"group-buy-market/internal/domain/trade/service/lock/factory"
	"group-buy-market/internal/types/enums"
	"group-buy-market/internal/types/exception"
)

type mockTradeRepo struct {
	activity *entity.GroupBuyActivityEntity
	count    int
	occupyOK bool
}

func (m *mockTradeRepo) QueryMarketPayOrderEntityByOutTradeNo(ctx context.Context, userID, outTradeNo string) (*entity.MarketPayOrderEntity, error) {
	return nil, nil
}
func (m *mockTradeRepo) LockMarketPayOrder(ctx context.Context, agg *aggregate.GroupBuyOrderAggregate) (*entity.MarketPayOrderEntity, error) {
	return nil, nil
}
func (m *mockTradeRepo) QueryGroupBuyProgress(ctx context.Context, teamID string) (*valobj.GroupBuyProgressVO, error) {
	return nil, nil
}
func (m *mockTradeRepo) QueryGroupBuyActivityEntityByActivityID(ctx context.Context, activityID int64) (*entity.GroupBuyActivityEntity, error) {
	return m.activity, nil
}
func (m *mockTradeRepo) QueryOrderCountByActivityID(ctx context.Context, activityID int64, userID string) (int, error) {
	return m.count, nil
}
func (m *mockTradeRepo) QueryGroupBuyTeamByTeamID(ctx context.Context, teamID string) (*entity.GroupBuyTeamEntity, error) {
	return nil, nil
}
func (m *mockTradeRepo) SettlementMarketPayOrder(ctx context.Context, agg *aggregate.GroupBuyTeamSettlementAggregate) (*entity.NotifyTaskEntity, error) {
	return nil, nil
}
func (m *mockTradeRepo) IsSCBlackIntercept(source, channel string) bool { return false }
func (m *mockTradeRepo) QueryUnExecutedNotifyTaskList(ctx context.Context) ([]*entity.NotifyTaskEntity, error) {
	return nil, nil
}
func (m *mockTradeRepo) QueryUnExecutedNotifyTaskListByTeamID(ctx context.Context, teamID string) ([]*entity.NotifyTaskEntity, error) {
	return nil, nil
}
func (m *mockTradeRepo) UpdateNotifyTaskStatusSuccess(ctx context.Context, task *entity.NotifyTaskEntity) (int, error) {
	return 0, nil
}
func (m *mockTradeRepo) UpdateNotifyTaskStatusError(ctx context.Context, task *entity.NotifyTaskEntity) (int, error) {
	return 0, nil
}
func (m *mockTradeRepo) UpdateNotifyTaskStatusRetry(ctx context.Context, task *entity.NotifyTaskEntity) (int, error) {
	return 0, nil
}
func (m *mockTradeRepo) OccupyTeamStock(ctx context.Context, teamStockKey, recoveryTeamStockKey string, target, validTime int) (bool, error) {
	return m.occupyOK, nil
}
func (m *mockTradeRepo) RecoveryTeamStock(ctx context.Context, recoveryTeamStockKey string, validTime int) error {
	return nil
}
func (m *mockTradeRepo) Unpaid2Refund(ctx context.Context, agg *aggregate.GroupBuyRefundAggregate) (*entity.NotifyTaskEntity, error) {
	return nil, nil
}
func (m *mockTradeRepo) Paid2Refund(ctx context.Context, agg *aggregate.GroupBuyRefundAggregate) (*entity.NotifyTaskEntity, error) {
	return nil, nil
}
func (m *mockTradeRepo) PaidTeam2Refund(ctx context.Context, agg *aggregate.GroupBuyRefundAggregate) (*entity.NotifyTaskEntity, error) {
	return nil, nil
}
func (m *mockTradeRepo) QueryTimeoutUnpaidOrderList(ctx context.Context) ([]*entity.TimeoutUnpaidOrderEntity, error) {
	return nil, nil
}
func (m *mockTradeRepo) RefundOrderExist(ctx context.Context, teamID, category, orderID string) (bool, error) {
	return false, nil
}
func (m *mockTradeRepo) Refund2AddRecovery(ctx context.Context, recoveryTeamStockKey, orderID string) error {
	return nil
}

func effectiveActivity() *entity.GroupBuyActivityEntity {
	return &entity.GroupBuyActivityEntity{
		ActivityID:     100123,
		TakeLimitCount: 2,
		Target:         3,
		ValidTime:      15,
		Status:         enums.ActivityEffective,
		StartTime:      time.Now().Add(-time.Hour),
		EndTime:        time.Now().Add(time.Hour),
	}
}

func TestLockChainOpenTeam(t *testing.T) {
	repo := &mockTradeRepo{activity: effectiveActivity(), count: 0, occupyOK: true}
	chain := NewTradeLockRuleFilter(repo)
	back, err := chain.Apply(context.Background(), entity.TradeLockRuleCommandEntity{
		UserID: "u1", ActivityID: 100123, TeamID: "",
	}, &factory.DynamicContext{})
	if err != nil {
		t.Fatal(err)
	}
	if back.UserTakeOrderCount != 0 {
		t.Fatalf("count %d", back.UserTakeOrderCount)
	}
}

func TestLockChainTakeLimit(t *testing.T) {
	repo := &mockTradeRepo{activity: effectiveActivity(), count: 2}
	chain := NewTradeLockRuleFilter(repo)
	_, err := chain.Apply(context.Background(), entity.TradeLockRuleCommandEntity{
		UserID: "u1", ActivityID: 100123,
	}, &factory.DynamicContext{})
	ae, ok := exception.AsAppException(err)
	if !ok || ae.Code != enums.E0103.Code {
		t.Fatalf("want E0103 got %v", err)
	}
}

func TestLockChainStockFail(t *testing.T) {
	repo := &mockTradeRepo{activity: effectiveActivity(), count: 0, occupyOK: false}
	chain := NewTradeLockRuleFilter(repo)
	_, err := chain.Apply(context.Background(), entity.TradeLockRuleCommandEntity{
		UserID: "u1", ActivityID: 100123, TeamID: "12345678",
	}, &factory.DynamicContext{})
	ae, ok := exception.AsAppException(err)
	if !ok || ae.Code != enums.E0008.Code {
		t.Fatalf("want E0008 got %v", err)
	}
}

func TestLockChainActivityInvalid(t *testing.T) {
	act := effectiveActivity()
	act.Status = enums.ActivityCreate
	repo := &mockTradeRepo{activity: act}
	chain := NewTradeLockRuleFilter(repo)
	_, err := chain.Apply(context.Background(), entity.TradeLockRuleCommandEntity{
		UserID: "u1", ActivityID: 100123,
	}, &factory.DynamicContext{})
	ae, ok := exception.AsAppException(err)
	if !ok || ae.Code != enums.E0101.Code {
		t.Fatalf("want E0101 got %v", err)
	}
}

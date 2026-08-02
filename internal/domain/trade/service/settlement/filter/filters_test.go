package filter

import (
	"context"
	"testing"
	"time"

	"group-buy-market/internal/domain/trade/model/aggregate"
	"group-buy-market/internal/domain/trade/model/entity"
	"group-buy-market/internal/domain/trade/model/valobj"
	"group-buy-market/internal/types/enums"
	"group-buy-market/internal/types/exception"
)

type mockSettleRepo struct {
	black    bool
	order    *entity.MarketPayOrderEntity
	team     *entity.GroupBuyTeamEntity
}

func (m *mockSettleRepo) QueryMarketPayOrderEntityByOutTradeNo(ctx context.Context, userID, outTradeNo string) (*entity.MarketPayOrderEntity, error) {
	return m.order, nil
}
func (m *mockSettleRepo) LockMarketPayOrder(ctx context.Context, agg *aggregate.GroupBuyOrderAggregate) (*entity.MarketPayOrderEntity, error) {
	return nil, nil
}
func (m *mockSettleRepo) QueryGroupBuyProgress(ctx context.Context, teamID string) (*valobj.GroupBuyProgressVO, error) {
	return nil, nil
}
func (m *mockSettleRepo) QueryGroupBuyActivityEntityByActivityID(ctx context.Context, activityID int64) (*entity.GroupBuyActivityEntity, error) {
	return nil, nil
}
func (m *mockSettleRepo) QueryOrderCountByActivityID(ctx context.Context, activityID int64, userID string) (int, error) {
	return 0, nil
}
func (m *mockSettleRepo) QueryGroupBuyTeamByTeamID(ctx context.Context, teamID string) (*entity.GroupBuyTeamEntity, error) {
	return m.team, nil
}
func (m *mockSettleRepo) SettlementMarketPayOrder(ctx context.Context, agg *aggregate.GroupBuyTeamSettlementAggregate) (*entity.NotifyTaskEntity, error) {
	return nil, nil
}
func (m *mockSettleRepo) IsSCBlackIntercept(source, channel string) bool { return m.black }
func (m *mockSettleRepo) QueryUnExecutedNotifyTaskList(ctx context.Context) ([]*entity.NotifyTaskEntity, error) {
	return nil, nil
}
func (m *mockSettleRepo) QueryUnExecutedNotifyTaskListByTeamID(ctx context.Context, teamID string) ([]*entity.NotifyTaskEntity, error) {
	return nil, nil
}
func (m *mockSettleRepo) UpdateNotifyTaskStatusSuccess(ctx context.Context, task *entity.NotifyTaskEntity) (int, error) {
	return 0, nil
}
func (m *mockSettleRepo) UpdateNotifyTaskStatusError(ctx context.Context, task *entity.NotifyTaskEntity) (int, error) {
	return 0, nil
}
func (m *mockSettleRepo) UpdateNotifyTaskStatusRetry(ctx context.Context, task *entity.NotifyTaskEntity) (int, error) {
	return 0, nil
}
func (m *mockSettleRepo) OccupyTeamStock(ctx context.Context, teamStockKey, recoveryTeamStockKey string, target, validTime int) (bool, error) {
	return true, nil
}
func (m *mockSettleRepo) RecoveryTeamStock(ctx context.Context, recoveryTeamStockKey string, validTime int) error {
	return nil
}
func (m *mockSettleRepo) Unpaid2Refund(ctx context.Context, agg *aggregate.GroupBuyRefundAggregate) (*entity.NotifyTaskEntity, error) {
	return nil, nil
}
func (m *mockSettleRepo) Paid2Refund(ctx context.Context, agg *aggregate.GroupBuyRefundAggregate) (*entity.NotifyTaskEntity, error) {
	return nil, nil
}
func (m *mockSettleRepo) PaidTeam2Refund(ctx context.Context, agg *aggregate.GroupBuyRefundAggregate) (*entity.NotifyTaskEntity, error) {
	return nil, nil
}
func (m *mockSettleRepo) QueryTimeoutUnpaidOrderList(ctx context.Context) ([]*entity.TimeoutUnpaidOrderEntity, error) {
	return nil, nil
}
func (m *mockSettleRepo) RefundOrderExist(ctx context.Context, teamID, category, orderID string) (bool, error) {
	return false, nil
}
func (m *mockSettleRepo) Refund2AddRecovery(ctx context.Context, recoveryTeamStockKey, orderID string) error {
	return nil
}

func TestSettlementSuccess(t *testing.T) {
	repo := &mockSettleRepo{
		order: &entity.MarketPayOrderEntity{TeamID: "t1", OrderID: "o1", TradeOrderStatus: valobj.TradeOrderCreate},
		team: &entity.GroupBuyTeamEntity{
			TeamID: "t1", ActivityID: 100123, TargetCount: 3, CompleteCount: 1, LockCount: 2,
			Status: enums.GroupBuyProgress, ValidEndTime: time.Now().Add(time.Hour),
		},
	}
	chain := NewTradeSettlementRuleFilter(repo)
	back, err := chain.Apply(context.Background(), entity.TradeSettlementRuleCommandEntity{
		Source: "s01", Channel: "c01", UserID: "u1", OutTradeNo: "n1", OutTradeTime: time.Now(),
	}, &SettleCtx{})
	if err != nil {
		t.Fatal(err)
	}
	if back.TeamID != "t1" || back.ActivityID != 100123 {
		t.Fatalf("%+v", back)
	}
}

func TestSettlementBlacklist(t *testing.T) {
	repo := &mockSettleRepo{black: true}
	chain := NewTradeSettlementRuleFilter(repo)
	_, err := chain.Apply(context.Background(), entity.TradeSettlementRuleCommandEntity{
		Source: "s02", Channel: "c02", UserID: "u", OutTradeNo: "n", OutTradeTime: time.Now(),
	}, &SettleCtx{})
	ae, ok := exception.AsAppException(err)
	if !ok || ae.Code != enums.E0105.Code {
		t.Fatalf("want E0105 got %v", err)
	}
}

func TestSettlementClosedOrder(t *testing.T) {
	repo := &mockSettleRepo{
		order: &entity.MarketPayOrderEntity{TradeOrderStatus: valobj.TradeOrderClose},
	}
	chain := NewTradeSettlementRuleFilter(repo)
	_, err := chain.Apply(context.Background(), entity.TradeSettlementRuleCommandEntity{
		Source: "s01", Channel: "c01", UserID: "u", OutTradeNo: "n", OutTradeTime: time.Now(),
	}, &SettleCtx{})
	ae, ok := exception.AsAppException(err)
	if !ok || ae.Code != enums.E0104.Code {
		t.Fatalf("want E0104 got %v", err)
	}
}

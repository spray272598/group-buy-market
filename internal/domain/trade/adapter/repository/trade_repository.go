package repository

import (
	"context"

	activityentity "group-buy-market/internal/domain/activity/model/entity"
	"group-buy-market/internal/domain/trade/model/aggregate"
	"group-buy-market/internal/domain/trade/model/entity"
	"group-buy-market/internal/domain/trade/model/valobj"
)

// ITradeRepository 交易仓储端口
type ITradeRepository interface {
	QueryMarketPayOrderEntityByOutTradeNo(ctx context.Context, userID, outTradeNo string) (*entity.MarketPayOrderEntity, error)
	LockMarketPayOrder(ctx context.Context, agg *aggregate.GroupBuyOrderAggregate) (*entity.MarketPayOrderEntity, error)
	QueryGroupBuyProgress(ctx context.Context, teamID string) (*valobj.GroupBuyProgressVO, error)
	QueryGroupBuyActivityEntityByActivityID(ctx context.Context, activityID int64) (*entity.GroupBuyActivityEntity, error)
	QueryOrderCountByActivityID(ctx context.Context, activityID int64, userID string) (int, error)
	QueryGroupBuyTeamByTeamID(ctx context.Context, teamID string) (*entity.GroupBuyTeamEntity, error)
	SettlementMarketPayOrder(ctx context.Context, agg *aggregate.GroupBuyTeamSettlementAggregate) (*entity.NotifyTaskEntity, error)
	IsSCBlackIntercept(source, channel string) bool
	QueryUnExecutedNotifyTaskList(ctx context.Context) ([]*entity.NotifyTaskEntity, error)
	QueryUnExecutedNotifyTaskListByTeamID(ctx context.Context, teamID string) ([]*entity.NotifyTaskEntity, error)
	UpdateNotifyTaskStatusSuccess(ctx context.Context, task *entity.NotifyTaskEntity) (int, error)
	UpdateNotifyTaskStatusError(ctx context.Context, task *entity.NotifyTaskEntity) (int, error)
	UpdateNotifyTaskStatusRetry(ctx context.Context, task *entity.NotifyTaskEntity) (int, error)
	OccupyTeamStock(ctx context.Context, teamStockKey, recoveryTeamStockKey string, target, validTime int) (bool, error)
	RecoveryTeamStock(ctx context.Context, recoveryTeamStockKey string, validTime int) error
	Unpaid2Refund(ctx context.Context, agg *aggregate.GroupBuyRefundAggregate) (*entity.NotifyTaskEntity, error)
	Paid2Refund(ctx context.Context, agg *aggregate.GroupBuyRefundAggregate) (*entity.NotifyTaskEntity, error)
	PaidTeam2Refund(ctx context.Context, agg *aggregate.GroupBuyRefundAggregate) (*entity.NotifyTaskEntity, error)
	QueryTimeoutUnpaidOrderList(ctx context.Context) ([]*activityentity.UserGroupBuyOrderDetailEntity, error)
	// RefundOrderExist 检查退单任务是否已存在（幂等）
	RefundOrderExist(ctx context.Context, teamID, category, orderID string) (bool, error)
	// Refund2AddRecovery 退单恢复 Redis 锁单库存（带 orderId 分布式锁防重）
	Refund2AddRecovery(ctx context.Context, recoveryTeamStockKey, orderID string) error
}

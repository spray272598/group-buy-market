package aggregate

import (
	"group-buy-market/internal/domain/trade/model/entity"
	"group-buy-market/internal/domain/trade/model/valobj"
)

// GroupBuyOrderAggregate 锁单聚合
type GroupBuyOrderAggregate struct {
	UserEntity         *entity.UserEntity
	PayActivityEntity  *entity.PayActivityEntity
	PayDiscountEntity  *entity.PayDiscountEntity
	UserTakeOrderCount int
}

// GroupBuyTeamSettlementAggregate 结算聚合
type GroupBuyTeamSettlementAggregate struct {
	UserEntity           *entity.UserEntity
	GroupBuyTeamEntity   *entity.GroupBuyTeamEntity
	TradePaySuccessEntity *entity.TradePaySuccessEntity
}

// GroupBuyRefundAggregate 退单聚合
type GroupBuyRefundAggregate struct {
	TradeRefundOrderEntity *entity.TradeRefundOrderEntity
	GroupBuyProgress       *valobj.GroupBuyProgressVO
}

func BuildUnpaid2RefundAggregate(order *entity.TradeRefundOrderEntity, lockDelta int) *GroupBuyRefundAggregate {
	return &GroupBuyRefundAggregate{
		TradeRefundOrderEntity: order,
		GroupBuyProgress: &valobj.GroupBuyProgressVO{
			LockCount: lockDelta, // 传 -1 表示锁单量-1
		},
	}
}

func BuildPaid2RefundAggregate(order *entity.TradeRefundOrderEntity, lockDelta, completeDelta int) *GroupBuyRefundAggregate {
	return &GroupBuyRefundAggregate{
		TradeRefundOrderEntity: order,
		GroupBuyProgress: &valobj.GroupBuyProgressVO{
			LockCount:     lockDelta,
			CompleteCount: completeDelta,
		},
	}
}

func BuildPaidTeam2RefundAggregate(order *entity.TradeRefundOrderEntity, lockDelta, completeDelta int) *GroupBuyRefundAggregate {
	return BuildPaid2RefundAggregate(order, lockDelta, completeDelta)
}

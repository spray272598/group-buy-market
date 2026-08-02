package lock

import (
	"context"
	"log/slog"

	"group-buy-market/internal/design/chain"
	"group-buy-market/internal/domain/trade/adapter/repository"
	"group-buy-market/internal/domain/trade/model/aggregate"
	"group-buy-market/internal/domain/trade/model/entity"
	"group-buy-market/internal/domain/trade/model/valobj"
	"group-buy-market/internal/domain/trade/service/lock/factory"
	"group-buy-market/internal/domain/trade/service/lock/filter"
)

// ITradeLockOrderService 锁单服务
type ITradeLockOrderService interface {
	QueryNoPayMarketPayOrderByOutTradeNo(ctx context.Context, userID, outTradeNo string) (*entity.MarketPayOrderEntity, error)
	QueryGroupBuyProgress(ctx context.Context, teamID string) (*valobj.GroupBuyProgressVO, error)
	LockMarketPayOrder(ctx context.Context, user *entity.UserEntity, activity *entity.PayActivityEntity, discount *entity.PayDiscountEntity) (*entity.MarketPayOrderEntity, error)
}

type TradeLockOrderService struct {
	repo   repository.ITradeRepository
	filter *chain.LinkedList[filter.LockReq, filter.LockCtx, *filter.LockRes]
}

func NewTradeLockOrderService(repo repository.ITradeRepository) *TradeLockOrderService {
	return &TradeLockOrderService{
		repo:   repo,
		filter: filter.NewTradeLockRuleFilter(repo),
	}
}

func (s *TradeLockOrderService) QueryNoPayMarketPayOrderByOutTradeNo(ctx context.Context, userID, outTradeNo string) (*entity.MarketPayOrderEntity, error) {
	slog.Info("拼团交易-查询未支付营销订单", "userId", userID, "outTradeNo", outTradeNo)
	return s.repo.QueryMarketPayOrderEntityByOutTradeNo(ctx, userID, outTradeNo)
}

func (s *TradeLockOrderService) QueryGroupBuyProgress(ctx context.Context, teamID string) (*valobj.GroupBuyProgressVO, error) {
	slog.Info("拼团交易-查询拼单进度", "teamId", teamID)
	return s.repo.QueryGroupBuyProgress(ctx, teamID)
}

func (s *TradeLockOrderService) LockMarketPayOrder(ctx context.Context, user *entity.UserEntity, activity *entity.PayActivityEntity, discount *entity.PayDiscountEntity) (*entity.MarketPayOrderEntity, error) {
	slog.Info("拼团交易-锁定营销优惠支付订单", "userId", user.UserID, "activityId", activity.ActivityID, "goodsId", discount.GoodsID)

	back, err := s.filter.Apply(ctx, entity.TradeLockRuleCommandEntity{
		ActivityID: activity.ActivityID,
		UserID:     user.UserID,
		TeamID:     activity.TeamID,
	}, &factory.DynamicContext{})
	if err != nil {
		return nil, err
	}

	agg := &aggregate.GroupBuyOrderAggregate{
		UserEntity:         user,
		PayActivityEntity:  activity,
		PayDiscountEntity:  discount,
		UserTakeOrderCount: back.UserTakeOrderCount,
	}

	order, err := s.repo.LockMarketPayOrder(ctx, agg)
	if err != nil {
		_ = s.repo.RecoveryTeamStock(ctx, back.RecoveryTeamStockKey, activity.ValidTime)
		return nil, err
	}
	return order, nil
}

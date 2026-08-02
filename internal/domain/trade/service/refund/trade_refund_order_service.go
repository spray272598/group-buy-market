package refund

import (
	"context"
	"log/slog"

	"group-buy-market/internal/design/chain"
	"group-buy-market/internal/domain/trade/adapter/repository"
	"group-buy-market/internal/domain/trade/model/entity"
	"group-buy-market/internal/domain/trade/model/valobj"
	"group-buy-market/internal/domain/trade/service/refund/business"
	"group-buy-market/internal/domain/trade/service/refund/filter"
)

// ITradeRefundOrderService 退单服务
type ITradeRefundOrderService interface {
	RefundOrder(ctx context.Context, cmd *entity.TradeRefundCommandEntity) (*entity.TradeRefundBehaviorEntity, error)
	RestoreTeamLockStock(ctx context.Context, msg *valobj.TeamRefundSuccess) error
	QueryTimeoutUnpaidOrderList(ctx context.Context) ([]*entity.TimeoutUnpaidOrderEntity, error)
}

type TradeRefundOrderService struct {
	repo       repository.ITradeRepository
	filter     *chain.LinkedList[filter.RefundReq, filter.RefundCtx, *filter.RefundRes]
	strategies business.StrategyMap
}

func NewTradeRefundOrderService(repo repository.ITradeRepository, strategies business.StrategyMap) *TradeRefundOrderService {
	return &TradeRefundOrderService{
		repo:       repo,
		filter:     filter.NewTradeRefundRuleFilter(repo, strategies),
		strategies: strategies,
	}
}

func (s *TradeRefundOrderService) RefundOrder(ctx context.Context, cmd *entity.TradeRefundCommandEntity) (*entity.TradeRefundBehaviorEntity, error) {
	slog.Info("逆向流程，退单操作", "userId", cmd.UserID, "outTradeNo", cmd.OutTradeNo)
	return s.filter.Apply(ctx, *cmd, &filter.RefundCtx{})
}

func (s *TradeRefundOrderService) RestoreTeamLockStock(ctx context.Context, msg *valobj.TeamRefundSuccess) error {
	slog.Info("逆向流程，恢复锁单量", "userId", msg.UserID, "activityId", msg.ActivityID, "teamId", msg.TeamID)
	rt, err := valobj.GetRefundTypeByCode(msg.Type)
	if err != nil {
		return err
	}
	strategy := s.strategies[rt.Strategy]
	if strategy == nil {
		return err
	}
	return strategy.ReverseStock(ctx, msg)
}

func (s *TradeRefundOrderService) QueryTimeoutUnpaidOrderList(ctx context.Context) ([]*entity.TimeoutUnpaidOrderEntity, error) {
	slog.Info("扫描数据，超时组队未支付订单")
	return s.repo.QueryTimeoutUnpaidOrderList(ctx)
}

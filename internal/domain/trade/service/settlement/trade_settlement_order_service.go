package settlement

import (
	"context"
	"log/slog"

	"group-buy-market/internal/design/chain"
	"group-buy-market/internal/domain/trade/adapter/repository"
	"group-buy-market/internal/domain/trade/model/aggregate"
	"group-buy-market/internal/domain/trade/model/entity"
	"group-buy-market/internal/domain/trade/service/settlement/filter"
	"group-buy-market/internal/domain/trade/service/task"
)

// ITradeSettlementOrderService 结算服务
type ITradeSettlementOrderService interface {
	SettlementMarketPayOrder(ctx context.Context, pay *entity.TradePaySuccessEntity) (*entity.TradePaySettlementEntity, error)
}

type TradeSettlementOrderService struct {
	repo   repository.ITradeRepository
	filter *chain.LinkedList[filter.SettleReq, filter.SettleCtx, *filter.SettleRes]
	task   *task.TradeTaskService
}

func NewTradeSettlementOrderService(repo repository.ITradeRepository, taskSvc *task.TradeTaskService) *TradeSettlementOrderService {
	return &TradeSettlementOrderService{
		repo:   repo,
		filter: filter.NewTradeSettlementRuleFilter(repo),
		task:   taskSvc,
	}
}

func (s *TradeSettlementOrderService) SettlementMarketPayOrder(ctx context.Context, pay *entity.TradePaySuccessEntity) (*entity.TradePaySettlementEntity, error) {
	slog.Info("拼团交易-支付订单结算", "userId", pay.UserID, "outTradeNo", pay.OutTradeNo)

	back, err := s.filter.Apply(ctx, entity.TradeSettlementRuleCommandEntity{
		Source:       pay.Source,
		Channel:      pay.Channel,
		UserID:       pay.UserID,
		OutTradeNo:   pay.OutTradeNo,
		OutTradeTime: pay.OutTradeTime,
	}, &filter.SettleCtx{})
	if err != nil {
		return nil, err
	}

	team := &entity.GroupBuyTeamEntity{
		TeamID:         back.TeamID,
		ActivityID:     back.ActivityID,
		TargetCount:    back.TargetCount,
		CompleteCount:  back.CompleteCount,
		LockCount:      back.LockCount,
		Status:         back.Status,
		ValidStartTime: back.ValidStartTime,
		ValidEndTime:   back.ValidEndTime,
		NotifyConfig:   back.NotifyConfig,
	}

	agg := &aggregate.GroupBuyTeamSettlementAggregate{
		UserEntity:            &entity.UserEntity{UserID: pay.UserID},
		GroupBuyTeamEntity:    team,
		TradePaySuccessEntity: pay,
	}

	notifyTask, err := s.repo.SettlementMarketPayOrder(ctx, agg)
	if err != nil {
		return nil, err
	}

	// 异步回调（失败由定时任务补偿）
	if notifyTask != nil && s.task != nil {
		go func() {
			result, e := s.task.ExecNotifyJob(context.Background(), notifyTask)
			if e != nil {
				slog.Error("回调通知拼团完结失败", "err", e, "result", result)
			} else {
				slog.Info("回调通知拼团完结", "result", result)
			}
		}()
	}

	return &entity.TradePaySettlementEntity{
		Source:     pay.Source,
		Channel:    pay.Channel,
		UserID:     pay.UserID,
		TeamID:     back.TeamID,
		ActivityID: team.ActivityID,
		OutTradeNo: pay.OutTradeNo,
	}, nil
}

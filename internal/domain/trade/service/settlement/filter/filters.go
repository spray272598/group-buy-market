package filter

import (
	"context"
	"log/slog"

	"group-buy-market/internal/design/chain"
	"group-buy-market/internal/domain/trade/adapter/repository"
	"group-buy-market/internal/domain/trade/model/entity"
	"group-buy-market/internal/domain/trade/model/valobj"
	"group-buy-market/internal/types/enums"
	"group-buy-market/internal/types/exception"
)

type SettleCtx struct {
	MarketPayOrder *entity.MarketPayOrderEntity
	GroupBuyTeam   *entity.GroupBuyTeamEntity
}

type SettleReq = entity.TradeSettlementRuleCommandEntity
type SettleRes = entity.TradeSettlementRuleFilterBackEntity

// SCRuleFilter SC 黑名单
type SCRuleFilter struct {
	chain.BaseHandler[SettleReq, SettleCtx, *SettleRes]
	Repo repository.ITradeRepository
}

func (f *SCRuleFilter) Apply(ctx context.Context, req SettleReq, dc *SettleCtx) (*SettleRes, error) {
	slog.Info("结算规则过滤-渠道黑名单校验", "userId", req.UserID, "outTradeNo", req.OutTradeNo)
	if f.Repo.IsSCBlackIntercept(req.Source, req.Channel) {
		return nil, exception.NewAppException(enums.E0105)
	}
	return f.Next(ctx, req, dc)
}

// OutTradeNoRuleFilter 外部单号校验
type OutTradeNoRuleFilter struct {
	chain.BaseHandler[SettleReq, SettleCtx, *SettleRes]
	Repo repository.ITradeRepository
}

func (f *OutTradeNoRuleFilter) Apply(ctx context.Context, req SettleReq, dc *SettleCtx) (*SettleRes, error) {
	slog.Info("结算规则过滤-外部单号校验", "userId", req.UserID, "outTradeNo", req.OutTradeNo)
	order, err := f.Repo.QueryMarketPayOrderEntityByOutTradeNo(ctx, req.UserID, req.OutTradeNo)
	if err != nil {
		return nil, err
	}
	if order == nil || order.TradeOrderStatus == valobj.TradeOrderClose {
		return nil, exception.NewAppException(enums.E0104)
	}
	dc.MarketPayOrder = order
	return f.Next(ctx, req, dc)
}

// SettableRuleFilter 有效时间校验
type SettableRuleFilter struct {
	chain.BaseHandler[SettleReq, SettleCtx, *SettleRes]
	Repo repository.ITradeRepository
}

func (f *SettableRuleFilter) Apply(ctx context.Context, req SettleReq, dc *SettleCtx) (*SettleRes, error) {
	slog.Info("结算规则过滤-有效时间校验", "userId", req.UserID, "outTradeNo", req.OutTradeNo)
	team, err := f.Repo.QueryGroupBuyTeamByTeamID(ctx, dc.MarketPayOrder.TeamID)
	if err != nil {
		return nil, err
	}
	if !req.OutTradeTime.Before(team.ValidEndTime) {
		return nil, exception.NewAppException(enums.E0106)
	}
	dc.GroupBuyTeam = team
	return f.Next(ctx, req, dc)
}

// EndRuleFilter 结束节点
type EndRuleFilter struct {
	chain.BaseHandler[SettleReq, SettleCtx, *SettleRes]
}

func (f *EndRuleFilter) Apply(ctx context.Context, req SettleReq, dc *SettleCtx) (*SettleRes, error) {
	slog.Info("结算规则过滤-结束节点", "userId", req.UserID, "outTradeNo", req.OutTradeNo)
	t := dc.GroupBuyTeam
	return &SettleRes{
		TeamID:         t.TeamID,
		ActivityID:     t.ActivityID,
		TargetCount:    t.TargetCount,
		CompleteCount:  t.CompleteCount,
		LockCount:      t.LockCount,
		Status:         t.Status,
		ValidStartTime: t.ValidStartTime,
		ValidEndTime:   t.ValidEndTime,
		NotifyConfig:   t.NotifyConfig,
	}, nil
}

func NewTradeSettlementRuleFilter(repo repository.ITradeRepository) *chain.LinkedList[SettleReq, SettleCtx, *SettleRes] {
	return chain.NewLinkedList[SettleReq, SettleCtx, *SettleRes](
		"交易结算规则过滤链",
		&SCRuleFilter{Repo: repo},
		&OutTradeNoRuleFilter{Repo: repo},
		&SettableRuleFilter{Repo: repo},
		&EndRuleFilter{},
	)
}

package filter

import (
	"context"
	"log/slog"

	"group-buy-market/internal/design/chain"
	"group-buy-market/internal/domain/trade/adapter/repository"
	"group-buy-market/internal/domain/trade/model/entity"
	"group-buy-market/internal/domain/trade/model/valobj"
	"group-buy-market/internal/domain/trade/service/refund/business"
)

type RefundCtx struct {
	MarketPayOrder *entity.MarketPayOrderEntity
	GroupBuyTeam   *entity.GroupBuyTeamEntity
}

type RefundReq = entity.TradeRefundCommandEntity
type RefundRes = entity.TradeRefundBehaviorEntity

// DataNodeFilter 加载数据
type DataNodeFilter struct {
	chain.BaseHandler[RefundReq, RefundCtx, *RefundRes]
	Repo repository.ITradeRepository
}

func (f *DataNodeFilter) Apply(ctx context.Context, req RefundReq, dc *RefundCtx) (*RefundRes, error) {
	slog.Info("逆向流程-退单操作，数据加载节点", "userId", req.UserID, "outTradeNo", req.OutTradeNo)
	order, err := f.Repo.QueryMarketPayOrderEntityByOutTradeNo(ctx, req.UserID, req.OutTradeNo)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, nil
	}
	team, err := f.Repo.QueryGroupBuyTeamByTeamID(ctx, order.TeamID)
	if err != nil {
		return nil, err
	}
	dc.MarketPayOrder = order
	dc.GroupBuyTeam = team
	return f.Next(ctx, req, dc)
}

// UniqueRefundNodeFilter 重复退单检查
type UniqueRefundNodeFilter struct {
	chain.BaseHandler[RefundReq, RefundCtx, *RefundRes]
}

func (f *UniqueRefundNodeFilter) Apply(ctx context.Context, req RefundReq, dc *RefundCtx) (*RefundRes, error) {
	slog.Info("逆向流程-退单操作，重复退单检查", "userId", req.UserID, "outTradeNo", req.OutTradeNo)
	if dc.MarketPayOrder != nil && dc.MarketPayOrder.TradeOrderStatus == valobj.TradeOrderClose {
		return &RefundRes{
			UserID:   req.UserID,
			OrderID:  dc.MarketPayOrder.OrderID,
			TeamID:   dc.MarketPayOrder.TeamID,
			Behavior: entity.RefundBehaviorRepeat,
		}, nil
	}
	return f.Next(ctx, req, dc)
}

// RefundOrderNodeFilter 执行退单策略
type RefundOrderNodeFilter struct {
	chain.BaseHandler[RefundReq, RefundCtx, *RefundRes]
	Strategies business.StrategyMap
}

func (f *RefundOrderNodeFilter) Apply(ctx context.Context, req RefundReq, dc *RefundCtx) (*RefundRes, error) {
	slog.Info("逆向流程-退单操作，退单策略处理", "userId", req.UserID, "outTradeNo", req.OutTradeNo)
	refundType, err := valobj.GetRefundStrategy(dc.GroupBuyTeam.Status, dc.MarketPayOrder.TradeOrderStatus)
	if err != nil {
		return nil, err
	}
	strategy := f.Strategies[refundType.Strategy]
	if strategy == nil {
		return nil, err
	}
	if err := strategy.RefundOrder(ctx, &entity.TradeRefundOrderEntity{
		UserID:     req.UserID,
		OrderID:    dc.MarketPayOrder.OrderID,
		TeamID:     dc.MarketPayOrder.TeamID,
		ActivityID: dc.GroupBuyTeam.ActivityID,
		OutTradeNo: req.OutTradeNo,
	}); err != nil {
		return nil, err
	}
	return &RefundRes{
		UserID:   req.UserID,
		OrderID:  dc.MarketPayOrder.OrderID,
		TeamID:   dc.MarketPayOrder.TeamID,
		Behavior: entity.RefundBehaviorSuccess,
	}, nil
}

func NewTradeRefundRuleFilter(repo repository.ITradeRepository, strategies business.StrategyMap) *chain.LinkedList[RefundReq, RefundCtx, *RefundRes] {
	return chain.NewLinkedList[RefundReq, RefundCtx, *RefundRes](
		"退单规则过滤链",
		&DataNodeFilter{Repo: repo},
		&UniqueRefundNodeFilter{},
		&RefundOrderNodeFilter{Strategies: strategies},
	)
}

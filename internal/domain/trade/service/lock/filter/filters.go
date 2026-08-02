package filter

import (
	"context"
	"log/slog"
	"time"

	"group-buy-market/internal/design/chain"
	"group-buy-market/internal/domain/trade/adapter/repository"
	"group-buy-market/internal/domain/trade/model/entity"
	"group-buy-market/internal/domain/trade/service/lock/factory"
	"group-buy-market/internal/types/enums"
	"group-buy-market/internal/types/exception"
)

type LockCtx = factory.DynamicContext
type LockReq = entity.TradeLockRuleCommandEntity
type LockRes = entity.TradeLockRuleFilterBackEntity

// ActivityUsabilityRuleFilter 活动可用性
type ActivityUsabilityRuleFilter struct {
	chain.BaseHandler[LockReq, LockCtx, *LockRes]
	Repo repository.ITradeRepository
}

func (f *ActivityUsabilityRuleFilter) Apply(ctx context.Context, req LockReq, dc *LockCtx) (*LockRes, error) {
	slog.Info("交易规则过滤-活动的可用性校验", "userId", req.UserID, "activityId", req.ActivityID)
	act, err := f.Repo.QueryGroupBuyActivityEntityByActivityID(ctx, req.ActivityID)
	if err != nil {
		return nil, err
	}
	if act == nil || act.Status != enums.ActivityEffective {
		return nil, exception.NewAppException(enums.E0101)
	}
	now := time.Now()
	if now.Before(act.StartTime) || now.After(act.EndTime) {
		return nil, exception.NewAppException(enums.E0102)
	}
	dc.GroupBuyActivity = act
	return f.Next(ctx, req, dc)
}

// UserTakeLimitRuleFilter 用户参与次数
type UserTakeLimitRuleFilter struct {
	chain.BaseHandler[LockReq, LockCtx, *LockRes]
	Repo repository.ITradeRepository
}

func (f *UserTakeLimitRuleFilter) Apply(ctx context.Context, req LockReq, dc *LockCtx) (*LockRes, error) {
	slog.Info("交易规则过滤-用户参与次数校验", "userId", req.UserID, "activityId", req.ActivityID)
	count, err := f.Repo.QueryOrderCountByActivityID(ctx, req.ActivityID, req.UserID)
	if err != nil {
		return nil, err
	}
	if dc.GroupBuyActivity != nil && dc.GroupBuyActivity.TakeLimitCount > 0 && count >= dc.GroupBuyActivity.TakeLimitCount {
		return nil, exception.NewAppException(enums.E0103)
	}
	dc.UserTakeOrderCount = count
	return f.Next(ctx, req, dc)
}

// TeamStockOccupyRuleFilter 组队库存占用
type TeamStockOccupyRuleFilter struct {
	chain.BaseHandler[LockReq, LockCtx, *LockRes]
	Repo repository.ITradeRepository
}

func (f *TeamStockOccupyRuleFilter) Apply(ctx context.Context, req LockReq, dc *LockCtx) (*LockRes, error) {
	slog.Info("交易规则过滤-组队库存校验", "userId", req.UserID, "activityId", req.ActivityID)
	if req.TeamID == "" {
		return &LockRes{UserTakeOrderCount: dc.UserTakeOrderCount}, nil
	}
	act := dc.GroupBuyActivity
	teamStockKey := dc.GenerateTeamStockKey(req.TeamID)
	recoveryKey := dc.GenerateRecoveryTeamStockKey(req.TeamID)
	ok, err := f.Repo.OccupyTeamStock(ctx, teamStockKey, recoveryKey, act.Target, act.ValidTime)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, exception.NewAppException(enums.E0008)
	}
	return &LockRes{
		UserTakeOrderCount:   dc.UserTakeOrderCount,
		RecoveryTeamStockKey: recoveryKey,
	}, nil
}

// NewTradeLockRuleFilter 组装锁单规则链
func NewTradeLockRuleFilter(repo repository.ITradeRepository) *chain.LinkedList[LockReq, LockCtx, *LockRes] {
	f1 := &ActivityUsabilityRuleFilter{Repo: repo}
	f2 := &UserTakeLimitRuleFilter{Repo: repo}
	f3 := &TeamStockOccupyRuleFilter{Repo: repo}
	return chain.NewLinkedList[LockReq, LockCtx, *LockRes]("交易规则过滤链", f1, f2, f3)
}

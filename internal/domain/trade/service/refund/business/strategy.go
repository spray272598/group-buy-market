package business

import (
	"context"
	"log/slog"

	"group-buy-market/internal/domain/trade/adapter/port"
	"group-buy-market/internal/domain/trade/adapter/repository"
	"group-buy-market/internal/domain/trade/model/aggregate"
	"group-buy-market/internal/domain/trade/model/entity"
	"group-buy-market/internal/domain/trade/model/valobj"
	"group-buy-market/internal/domain/trade/service/lock/factory"
	"group-buy-market/internal/domain/trade/service/task"
)

// IRefundOrderStrategy 退单策略
type IRefundOrderStrategy interface {
	Name() string
	RefundOrder(ctx context.Context, order *entity.TradeRefundOrderEntity) error
	ReverseStock(ctx context.Context, msg *valobj.TeamRefundSuccess) error
}

type baseStrategy struct {
	Repo repository.ITradeRepository
	Port port.ITradePort
	Task *task.TradeTaskService
}

func (b *baseStrategy) sendRefundNotify(ctx context.Context, notify *entity.NotifyTaskEntity, scene string) {
	if notify == nil || b.Task == nil {
		return
	}
	go func() {
		result, err := b.Task.ExecNotifyJob(context.Background(), notify)
		if err != nil {
			slog.Error("退单回调失败", "scene", scene, "err", err, "result", result)
		} else {
			slog.Info("退单回调完成", "scene", scene, "result", result)
		}
	}()
}

func (b *baseStrategy) doReverseStock(ctx context.Context, msg *valobj.TeamRefundSuccess, scene string) error {
	slog.Info("逆向库存恢复", "scene", scene, "userId", msg.UserID, "teamId", msg.TeamID, "orderId", msg.OrderID)
	key := factory.GenerateRecoveryTeamStockKey(msg.ActivityID, msg.TeamID)
	// 对齐 Java：refund2AddRecovery（orderId 防重 + recovery incr）
	return b.Repo.Refund2AddRecovery(ctx, key, msg.OrderID)
}

// Unpaid2RefundStrategy 未支付未成团
type Unpaid2RefundStrategy struct {
	baseStrategy
}

func NewUnpaid2Refund(repo repository.ITradeRepository, p port.ITradePort, t *task.TradeTaskService) *Unpaid2RefundStrategy {
	return &Unpaid2RefundStrategy{baseStrategy: baseStrategy{Repo: repo, Port: p, Task: t}}
}

func (s *Unpaid2RefundStrategy) Name() string { return "unpaid2RefundStrategy" }

func (s *Unpaid2RefundStrategy) RefundOrder(ctx context.Context, order *entity.TradeRefundOrderEntity) error {
	slog.Info("退单；未支付，未成团", "userId", order.UserID, "teamId", order.TeamID, "orderId", order.OrderID)
	notify, err := s.Repo.Unpaid2Refund(ctx, aggregate.BuildUnpaid2RefundAggregate(order, -1))
	if err != nil {
		return err
	}
	s.sendRefundNotify(ctx, notify, "未支付，未成团")
	return nil
}

func (s *Unpaid2RefundStrategy) ReverseStock(ctx context.Context, msg *valobj.TeamRefundSuccess) error {
	return s.doReverseStock(ctx, msg, "未支付，未成团")
}

// Paid2RefundStrategy 已支付未成团
type Paid2RefundStrategy struct {
	baseStrategy
}

func NewPaid2Refund(repo repository.ITradeRepository, p port.ITradePort, t *task.TradeTaskService) *Paid2RefundStrategy {
	return &Paid2RefundStrategy{baseStrategy: baseStrategy{Repo: repo, Port: p, Task: t}}
}

func (s *Paid2RefundStrategy) Name() string { return "paid2RefundStrategy" }

func (s *Paid2RefundStrategy) RefundOrder(ctx context.Context, order *entity.TradeRefundOrderEntity) error {
	slog.Info("退单；已支付，未成团", "userId", order.UserID, "teamId", order.TeamID, "orderId", order.OrderID)
	notify, err := s.Repo.Paid2Refund(ctx, aggregate.BuildPaid2RefundAggregate(order, -1, -1))
	if err != nil {
		return err
	}
	s.sendRefundNotify(ctx, notify, "已支付，未成团")
	return nil
}

func (s *Paid2RefundStrategy) ReverseStock(ctx context.Context, msg *valobj.TeamRefundSuccess) error {
	return s.doReverseStock(ctx, msg, "已支付，未成团")
}

// PaidTeam2RefundStrategy 已支付已成团
type PaidTeam2RefundStrategy struct {
	baseStrategy
}

func NewPaidTeam2Refund(repo repository.ITradeRepository, p port.ITradePort, t *task.TradeTaskService) *PaidTeam2RefundStrategy {
	return &PaidTeam2RefundStrategy{baseStrategy: baseStrategy{Repo: repo, Port: p, Task: t}}
}

func (s *PaidTeam2RefundStrategy) Name() string { return "paidTeam2RefundStrategy" }

func (s *PaidTeam2RefundStrategy) RefundOrder(ctx context.Context, order *entity.TradeRefundOrderEntity) error {
	slog.Info("退单；已支付，已成团", "userId", order.UserID, "teamId", order.TeamID, "orderId", order.OrderID)
	notify, err := s.Repo.PaidTeam2Refund(ctx, aggregate.BuildPaidTeam2RefundAggregate(order, -1, -1))
	if err != nil {
		return err
	}
	s.sendRefundNotify(ctx, notify, "已支付，已成团")
	return nil
}

func (s *PaidTeam2RefundStrategy) ReverseStock(ctx context.Context, msg *valobj.TeamRefundSuccess) error {
	return s.doReverseStock(ctx, msg, "已支付，已成团")
}

// StrategyMap 策略注册表
type StrategyMap map[string]IRefundOrderStrategy

func NewStrategyMap(strategies ...IRefundOrderStrategy) StrategyMap {
	m := make(StrategyMap)
	for _, s := range strategies {
		m[s.Name()] = s
	}
	return m
}

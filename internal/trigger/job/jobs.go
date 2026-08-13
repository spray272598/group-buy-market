package job

import (
	"context"
	"log/slog"
	"time"

	"group-buy-market/internal/domain/trade/model/entity"
	"group-buy-market/internal/domain/trade/service/refund"
	"group-buy-market/internal/domain/trade/service/task"
	"group-buy-market/internal/infrastructure/metrics"
	redisx "group-buy-market/internal/infrastructure/redis"
	"group-buy-market/internal/types/safego"
)

// NotifyJob 回调补偿定时任务（多实例 Redis 分布式锁）
type NotifyJob struct {
	task     *task.TradeTaskService
	redis    *redisx.Service
	interval time.Duration
	stop     chan struct{}
}

func NewNotifyJob(taskSvc *task.TradeTaskService, rdb *redisx.Service, intervalSec int) *NotifyJob {
	if intervalSec <= 0 {
		intervalSec = 30
	}
	return &NotifyJob{
		task:     taskSvc,
		redis:    rdb,
		interval: time.Duration(intervalSec) * time.Second,
		stop:     make(chan struct{}),
	}
}

func (j *NotifyJob) Start() {
	go func() {
		ticker := time.NewTicker(j.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				j.exec()
			case <-j.stop:
				return
			}
		}
	}()
}

func (j *NotifyJob) exec() {
	defer safego.Recover("notify_job")
	start := time.Now()
	ctx := context.Background()
	lockKey := "group_buy_market_notify_job_exec"
	ok, err := j.redis.TryLockWait(ctx, lockKey, 3*time.Second, 50*time.Second)
	if err != nil || !ok {
		metrics.ObserveJob("notify", "skip_lock")
		return
	}
	defer func() { _ = j.redis.Unlock(context.Background(), lockKey) }()

	result, err := j.task.ExecNotifyJob(ctx, nil)
	metrics.ObserveJobDuration("notify", time.Since(start).Seconds())
	if err != nil {
		metrics.ObserveJob("notify", "error")
		slog.Error("定时任务，回调通知失败", "err", err)
		return
	}
	metrics.ObserveJob("notify", "success")
	if n := result["waitCount"]; n > 0 {
		metrics.ObserveJobProcessed("notify", "wait", n)
		slog.Info("定时任务，回调通知完成", "result", result)
	}
}

func (j *NotifyJob) Stop() { close(j.stop) }

// TimeoutRefundJob 超时未支付退单扫描（多实例分布式锁）
type TimeoutRefundJob struct {
	refund   refund.ITradeRefundOrderService
	redis    *redisx.Service
	interval time.Duration
	stop     chan struct{}
}

func NewTimeoutRefundJob(refundSvc refund.ITradeRefundOrderService, rdb *redisx.Service, intervalSec int) *TimeoutRefundJob {
	if intervalSec <= 0 {
		intervalSec = 60
	}
	return &TimeoutRefundJob{
		refund:   refundSvc,
		redis:    rdb,
		interval: time.Duration(intervalSec) * time.Second,
		stop:     make(chan struct{}),
	}
}

func (j *TimeoutRefundJob) Start() {
	go func() {
		ticker := time.NewTicker(j.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				j.exec()
			case <-j.stop:
				return
			}
		}
	}()
}

func (j *TimeoutRefundJob) exec() {
	defer safego.Recover("timeout_refund_job")
	start := time.Now()
	ctx := context.Background()
	lockKey := "group_buy_market_timeout_refund_job_exec"
	ok, err := j.redis.TryLockWait(ctx, lockKey, 3*time.Second, 60*time.Second)
	if err != nil || !ok {
		metrics.ObserveJob("timeout_refund", "skip_lock")
		slog.Info("超时退单定时任务，获取锁失败，跳过")
		return
	}
	defer func() { _ = j.redis.Unlock(context.Background(), lockKey) }()

	slog.Info("超时退单定时任务开始执行")
	list, err := j.refund.QueryTimeoutUnpaidOrderList(ctx)
	if err != nil {
		metrics.ObserveJobDuration("timeout_refund", time.Since(start).Seconds())
		metrics.ObserveJob("timeout_refund", "error")
		slog.Error("TimeoutRefundJob 扫描失败", "err", err)
		return
	}
	if len(list) == 0 {
		metrics.ObserveJobDuration("timeout_refund", time.Since(start).Seconds())
		metrics.ObserveJob("timeout_refund", "success")
		slog.Info("超时退单定时任务，未发现超时未支付订单")
		return
	}

	success, fail := 0, 0
	for _, item := range list {
		src, ch := item.Source, item.Channel
		if src == "" {
			src = "s01"
		}
		if ch == "" {
			ch = "c01"
		}
		_, err := j.refund.RefundOrder(ctx, &entity.TradeRefundCommandEntity{
			UserID:     item.UserID,
			OutTradeNo: item.OutTradeNo,
			Source:     src,
			Channel:    ch,
		})
		if err != nil {
			fail++
			slog.Error("超时订单退单失败", "userId", item.UserID, "outTradeNo", item.OutTradeNo, "err", err)
		} else {
			success++
			slog.Info("超时订单退单成功", "userId", item.UserID, "outTradeNo", item.OutTradeNo)
		}
	}
	metrics.ObserveJobDuration("timeout_refund", time.Since(start).Seconds())
	metrics.ObserveJob("timeout_refund", "success")
	metrics.ObserveJobProcessed("timeout_refund", "success", success)
	metrics.ObserveJobProcessed("timeout_refund", "fail", fail)
	slog.Info("超时退单定时任务执行完成", "success", success, "fail", fail)
}

func (j *TimeoutRefundJob) Stop() { close(j.stop) }

package notify

import (
	"context"
	"log/slog"
	"time"

	"group-buy-market/internal/domain/trade/adapter/port"
	"group-buy-market/internal/domain/trade/model/entity"
	"group-buy-market/internal/domain/trade/model/valobj"
	"group-buy-market/internal/infrastructure/gateway"
	"group-buy-market/internal/infrastructure/mq"
	redisx "group-buy-market/internal/infrastructure/redis"
	"group-buy-market/internal/types/enums"
)

// TradePort 实现 ITradePort：多实例抢锁 + HTTP/MQ 回调
// 对齐 Java infrastructure.adapter.port.TradePort
type TradePort struct {
	redis   *redisx.Service
	httpGW  *gateway.GroupBuyNotifyService
	publish *mq.EventPublisher
}

func NewTradePort(rdb *redisx.Service, httpGW *gateway.GroupBuyNotifyService, pub *mq.EventPublisher) port.ITradePort {
	return &TradePort{redis: rdb, httpGW: httpGW, publish: pub}
}

func (p *TradePort) GroupBuyNotify(ctx context.Context, task *entity.NotifyTaskEntity) (string, error) {
	if task == nil {
		return enums.NotifyTaskHTTPSuccess, nil
	}
	lockKey := task.LockKey()
	// 多实例部署：同一回调任务只允许一个实例执行
	ok, err := p.redis.TryLockWait(ctx, lockKey, 3*time.Second, 30*time.Second)
	if err != nil {
		slog.Error("回调任务抢锁异常", "lockKey", lockKey, "err", err)
		return enums.NotifyTaskHTTPNull, nil
	}
	if !ok {
		slog.Info("回调任务未抢到锁，跳过", "lockKey", lockKey)
		return enums.NotifyTaskHTTPNull, nil
	}
	defer func() { _ = p.redis.Unlock(context.Background(), lockKey) }()

	switch valobj.NotifyType(task.NotifyType) {
	case valobj.NotifyHTTP:
		if task.NotifyUrl == "" || task.NotifyUrl == "暂无" {
			return enums.NotifyTaskHTTPSuccess, nil
		}
		return p.httpGW.GroupBuyNotify(ctx, task.NotifyUrl, task.ParameterJSON)
	case valobj.NotifyMQ:
		routingKey := task.NotifyMQ
		if routingKey == "" {
			routingKey = "topic.team_success"
		}
		if err := p.publish.Publish(ctx, routingKey, task.ParameterJSON); err != nil {
			return enums.NotifyTaskHTTPError, err
		}
		return enums.NotifyTaskHTTPSuccess, nil
	default:
		slog.Warn("未知回调类型", "type", task.NotifyType)
		return enums.NotifyTaskHTTPSuccess, nil
	}
}

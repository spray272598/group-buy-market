package acl

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

// TradeNotifyACL 出站防腐层实现（Anti-Corruption Layer）
//
// 将领域 NotifyTask 翻译为：
//   - HTTP JSON 回调（下游商城协议）
//   - RabbitMQ 持久化消息
// 并处理多实例分布式锁，隔离领域与中间件细节。
// 对齐 Java infrastructure.adapter.port.TradePort
type TradeNotifyACL struct {
	redis   *redisx.Service
	httpGW  *gateway.GroupBuyNotifyService
	publish *mq.EventPublisher
}

// NewTradeNotifyACL 创建出站防腐实现，对外仍表现为 ITradePort
func NewTradeNotifyACL(rdb *redisx.Service, httpGW *gateway.GroupBuyNotifyService, pub *mq.EventPublisher) port.ITradePort {
	return &TradeNotifyACL{redis: rdb, httpGW: httpGW, publish: pub}
}

// 兼容旧名称
func NewTradePort(rdb *redisx.Service, httpGW *gateway.GroupBuyNotifyService, pub *mq.EventPublisher) port.ITradePort {
	return NewTradeNotifyACL(rdb, httpGW, pub)
}

func (p *TradeNotifyACL) GroupBuyNotify(ctx context.Context, task *entity.NotifyTaskEntity) (string, error) {
	if task == nil {
		return enums.NotifyTaskHTTPSuccess, nil
	}
	lockKey := task.LockKey()
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
		// ACL：领域参数 → HTTP 网关协议
		return p.httpGW.GroupBuyNotify(ctx, task.NotifyUrl, task.ParameterJSON)
	case valobj.NotifyMQ:
		routingKey := task.NotifyMQ
		if routingKey == "" {
			routingKey = "topic.team_success"
		}
		// ACL：领域参数 → MQ routingKey + body
		if err := p.publish.Publish(ctx, routingKey, task.ParameterJSON); err != nil {
			return enums.NotifyTaskHTTPError, err
		}
		return enums.NotifyTaskHTTPSuccess, nil
	default:
		slog.Warn("未知回调类型", "type", task.NotifyType)
		return enums.NotifyTaskHTTPSuccess, nil
	}
}

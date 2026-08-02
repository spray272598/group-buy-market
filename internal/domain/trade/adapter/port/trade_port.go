package port

import (
	"context"

	"group-buy-market/internal/domain/trade/model/entity"
)

// ITradePort 拼团结果通知出站端口（出站防腐层接口）
//
// 外部系统可能是：商城 HTTP 回调、MQ 消息总线等。
// 领域只表达「需要通知」，不表达「如何 HTTP/如何序列化」。
type ITradePort interface {
	// GroupBuyNotify 执行一次回调任务；返回 success / error / null
	GroupBuyNotify(ctx context.Context, task *entity.NotifyTaskEntity) (string, error)
}

package port

import (
	"context"

	"group-buy-market/internal/domain/trade/model/entity"
)

// ITradePort 交易端口：回调通知（HTTP/MQ）
type ITradePort interface {
	GroupBuyNotify(ctx context.Context, task *entity.NotifyTaskEntity) (string, error)
}

package listener

import (
	"context"
	"encoding/json"
	"log/slog"

	"group-buy-market/internal/domain/trade/model/valobj"
	"group-buy-market/internal/domain/trade/service/refund"
	"group-buy-market/internal/infrastructure/mq"
)

// StartConsumers 启动 RabbitMQ 消费者（对齐 Java @RabbitListener）
func StartConsumers(client *mq.Client, refundSvc refund.ITradeRefundOrderService) error {
	if client == nil {
		slog.Warn("RabbitMQ 未连接，跳过消费者启动")
		return nil
	}

	// 组队成功消息（外部业务可监听，这里仅日志）
	if err := client.Consume(client.TeamSuccessQueue(), func(body []byte) error {
		slog.Info("接收消息（组队成功）", "message", string(body))
		return nil
	}); err != nil {
		return err
	}

	// 退单成功 → 恢复锁单库存（最终一致性）
	if err := client.Consume(client.TeamRefundQueue(), func(body []byte) error {
		slog.Info("接收消息（退单成功）- 恢复拼团队伍锁单量", "message", string(body))
		var msg valobj.TeamRefundSuccess
		if err := json.Unmarshal(body, &msg); err != nil {
			return err
		}
		if err := refundSvc.RestoreTeamLockStock(context.Background(), &msg); err != nil {
			slog.Error("恢复锁单库存失败", "err", err)
			return err // Nack 重试
		}
		return nil
	}); err != nil {
		return err
	}

	slog.Info("RabbitMQ 消费者已启动",
		"successQueue", client.TeamSuccessQueue(),
		"refundQueue", client.TeamRefundQueue())
	return nil
}

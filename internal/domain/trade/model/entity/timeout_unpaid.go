package entity

import "time"

// TimeoutUnpaidOrderEntity 超时未支付订单（交易上下文自有模型，不依赖 activity 实体）
// 用于 TimeoutRefundJob 扫描与逆向退单编排。
type TimeoutUnpaidOrderEntity struct {
	UserID         string
	TeamID         string
	ActivityID     int64
	TargetCount    int
	CompleteCount  int
	LockCount      int
	ValidStartTime time.Time
	ValidEndTime   time.Time
	OutTradeNo     string
	Source         string
	Channel        string
}

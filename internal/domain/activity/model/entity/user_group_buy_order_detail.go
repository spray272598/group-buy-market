package entity

import "time"

// UserGroupBuyOrderDetailEntity 用户拼团明细（活动上下文：首页展示用）
// 超时未支付扫描请用 trade 上下文 TimeoutUnpaidOrderEntity，避免跨上下文污染。
type UserGroupBuyOrderDetailEntity struct {
	UserID         string
	TeamID         string
	ActivityID     int64
	TargetCount    int
	CompleteCount  int
	LockCount      int
	ValidStartTime time.Time
	ValidEndTime   time.Time
	OutTradeNo     string
}

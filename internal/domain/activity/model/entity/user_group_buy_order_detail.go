package entity

import "time"

// UserGroupBuyOrderDetailEntity 用户拼团明细（首页展示 / 超时退单扫描）
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
	Source         string
	Channel        string
}

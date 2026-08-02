package valobj

// NotifyType 回调类型
type NotifyType string

const (
	NotifyHTTP NotifyType = "HTTP"
	NotifyMQ   NotifyType = "MQ"
)

// NotifyConfigVO 回调配置
type NotifyConfigVO struct {
	NotifyType NotifyType
	NotifyMQ   string
	NotifyUrl  string
}

// TaskNotifyCategory 回调种类
type TaskNotifyCategory string

const (
	TaskTradeSettlement   TaskNotifyCategory = "trade_settlement"
	TaskTradeUnpaid2Refund TaskNotifyCategory = "trade_unpaid2refund"
	TaskTradePaid2Refund   TaskNotifyCategory = "trade_paid2refund"
	TaskTradePaidTeam2Refund TaskNotifyCategory = "trade_paid_team2refund"
)

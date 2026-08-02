package valobj

// TeamRefundSuccess 退单成功消息体（恢复库存用）
type TeamRefundSuccess struct {
	Type       string `json:"type"`
	UserID     string `json:"userId"`
	TeamID     string `json:"teamId"`
	OrderID    string `json:"orderId"`
	OutTradeNo string `json:"outTradeNo"`
	ActivityID int64  `json:"activityId"`
}

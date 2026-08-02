package valobj

// TradeOrderStatus 订单明细状态 0初始锁定、1消费完成、2用户退单
type TradeOrderStatus int

const (
	TradeOrderCreate   TradeOrderStatus = 0
	TradeOrderComplete TradeOrderStatus = 1
	TradeOrderClose    TradeOrderStatus = 2
)

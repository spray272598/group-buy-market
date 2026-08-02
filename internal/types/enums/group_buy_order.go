package enums

// GroupBuyOrderStatus 拼团组队状态
// 0-拼单中、1-完成、2-失败、3-完成-含退单
type GroupBuyOrderStatus int

const (
	GroupBuyProgress     GroupBuyOrderStatus = 0
	GroupBuyComplete     GroupBuyOrderStatus = 1
	GroupBuyFail         GroupBuyOrderStatus = 2
	GroupBuyCompleteFail GroupBuyOrderStatus = 3 // 完成-含退单
)

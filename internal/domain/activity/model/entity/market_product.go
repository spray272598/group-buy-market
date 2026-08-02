package entity

// MarketProductEntity 营销商品试算请求实体
type MarketProductEntity struct {
	UserID     string
	GoodsID    string
	Source     string
	Channel    string
	ActivityID *int64 // 可空，空则按 SC+goods 查询
}

package valobj

import "github.com/shopspring/decimal"

// SkuVO 商品值对象
type SkuVO struct {
	GoodsID       string
	GoodsName     string
	OriginalPrice decimal.Decimal
}

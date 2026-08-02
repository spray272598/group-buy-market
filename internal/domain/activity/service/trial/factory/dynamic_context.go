package factory

import (
	"github.com/shopspring/decimal"

	"group-buy-market/internal/domain/activity/model/valobj"
)

// DynamicContext 试算责任链/策略树动态上下文
type DynamicContext struct {
	GroupBuyActivityDiscountVO *valobj.GroupBuyActivityDiscountVO
	SkuVO                      *valobj.SkuVO
	DeductionPrice             *decimal.Decimal
	PayPrice                   *decimal.Decimal
	Visible                    bool
	Enable                     bool
}

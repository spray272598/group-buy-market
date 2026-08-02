package entity

import (
	"time"

	"github.com/shopspring/decimal"

	"group-buy-market/internal/domain/activity/model/valobj"
)

// TrialBalanceEntity 优惠试算结果
type TrialBalanceEntity struct {
	GoodsID                   string
	GoodsName                 string
	OriginalPrice             decimal.Decimal
	DeductionPrice            decimal.Decimal
	PayPrice                  decimal.Decimal
	TargetCount               int
	StartTime                 time.Time
	EndTime                   time.Time
	IsVisible                 bool
	IsEnable                  bool
	GroupBuyActivityDiscountVO *valobj.GroupBuyActivityDiscountVO
}

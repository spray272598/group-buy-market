package discount

import (
	"context"

	"github.com/shopspring/decimal"

	"group-buy-market/internal/domain/activity/adapter/repository"
	"group-buy-market/internal/domain/activity/model/valobj"
)

// ZJCalculateService 直减
type ZJCalculateService struct {
	BaseDiscount
}

func NewZJ(repo repository.IActivityRepository) *ZJCalculateService {
	return &ZJCalculateService{BaseDiscount: BaseDiscount{Repo: repo}}
}

func (s *ZJCalculateService) Plan() string { return "ZJ" }

func (s *ZJCalculateService) Calculate(ctx context.Context, userID string, originalPrice decimal.Decimal, discount *valobj.GroupBuyDiscount) (decimal.Decimal, error) {
	return s.FilterAndCalculate(ctx, userID, originalPrice, discount, func(op decimal.Decimal, d *valobj.GroupBuyDiscount) decimal.Decimal {
		expr, _ := decimal.NewFromString(d.MarketExpr)
		return minPay(op.Sub(expr))
	})
}

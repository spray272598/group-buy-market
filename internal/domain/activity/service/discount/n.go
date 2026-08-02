package discount

import (
	"context"

	"github.com/shopspring/decimal"

	"group-buy-market/internal/domain/activity/adapter/repository"
	"group-buy-market/internal/domain/activity/model/valobj"
)

// NCalculateService N元购
type NCalculateService struct {
	BaseDiscount
}

func NewN(repo repository.IActivityRepository) *NCalculateService {
	return &NCalculateService{BaseDiscount: BaseDiscount{Repo: repo}}
}

func (s *NCalculateService) Plan() string { return "N" }

func (s *NCalculateService) Calculate(ctx context.Context, userID string, originalPrice decimal.Decimal, discount *valobj.GroupBuyDiscount) (decimal.Decimal, error) {
	return s.FilterAndCalculate(ctx, userID, originalPrice, discount, func(op decimal.Decimal, d *valobj.GroupBuyDiscount) decimal.Decimal {
		expr, _ := decimal.NewFromString(d.MarketExpr)
		return expr
	})
}

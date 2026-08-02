package discount

import (
	"context"
	"strings"

	"github.com/shopspring/decimal"

	"group-buy-market/internal/domain/activity/adapter/repository"
	"group-buy-market/internal/domain/activity/model/valobj"
	"group-buy-market/internal/types/common"
)

// MJCalculateService 满减 表达式: 100,10 满100减10
type MJCalculateService struct {
	BaseDiscount
}

func NewMJ(repo repository.IActivityRepository) *MJCalculateService {
	return &MJCalculateService{BaseDiscount: BaseDiscount{Repo: repo}}
}

func (s *MJCalculateService) Plan() string { return "MJ" }

func (s *MJCalculateService) Calculate(ctx context.Context, userID string, originalPrice decimal.Decimal, discount *valobj.GroupBuyDiscount) (decimal.Decimal, error) {
	return s.FilterAndCalculate(ctx, userID, originalPrice, discount, func(op decimal.Decimal, d *valobj.GroupBuyDiscount) decimal.Decimal {
		parts := strings.Split(d.MarketExpr, common.Split)
		if len(parts) < 2 {
			return op
		}
		x, _ := decimal.NewFromString(strings.TrimSpace(parts[0]))
		y, _ := decimal.NewFromString(strings.TrimSpace(parts[1]))
		if op.LessThan(x) {
			return op
		}
		return minPay(op.Sub(y))
	})
}

package discount

import (
	"context"

	"github.com/shopspring/decimal"

	"group-buy-market/internal/domain/activity/adapter/repository"
	"group-buy-market/internal/domain/activity/model/valobj"
)

// ZKCalculateService 折扣 表达式为折扣比例 如 0.8
type ZKCalculateService struct {
	BaseDiscount
}

func NewZK(repo repository.IActivityRepository) *ZKCalculateService {
	return &ZKCalculateService{BaseDiscount: BaseDiscount{Repo: repo}}
}

func (s *ZKCalculateService) Plan() string { return "ZK" }

func (s *ZKCalculateService) Calculate(ctx context.Context, userID string, originalPrice decimal.Decimal, discount *valobj.GroupBuyDiscount) (decimal.Decimal, error) {
	return s.FilterAndCalculate(ctx, userID, originalPrice, discount, func(op decimal.Decimal, d *valobj.GroupBuyDiscount) decimal.Decimal {
		expr, _ := decimal.NewFromString(d.MarketExpr)
		// 向下取整到分... Java 用 setScale(0, DOWN) 取整到元
		price := op.Mul(expr).Truncate(0)
		return minPay(price)
	})
}

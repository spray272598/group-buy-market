package discount

import (
	"context"
	"log/slog"

	"github.com/shopspring/decimal"

	"group-buy-market/internal/domain/activity/adapter/repository"
	"group-buy-market/internal/domain/activity/model/valobj"
)

// IDiscountCalculateService 折扣计算策略接口
type IDiscountCalculateService interface {
	// Plan 策略编码 ZJ/MJ/N/ZK
	Plan() string
	Calculate(ctx context.Context, userID string, originalPrice decimal.Decimal, discount *valobj.GroupBuyDiscount) (decimal.Decimal, error)
}

// BaseDiscount 抽象折扣：人群标签过滤 + doCalculate
type BaseDiscount struct {
	Repo repository.IActivityRepository
}

func (b *BaseDiscount) FilterAndCalculate(
	ctx context.Context,
	userID string,
	originalPrice decimal.Decimal,
	discount *valobj.GroupBuyDiscount,
	doCalc func(originalPrice decimal.Decimal, discount *valobj.GroupBuyDiscount) decimal.Decimal,
) (decimal.Decimal, error) {
	if discount.DiscountType == valobj.DiscountTypeTag {
		ok, err := b.Repo.IsTagCrowdRange(ctx, discount.TagID, userID)
		if err != nil {
			return decimal.Zero, err
		}
		if !ok {
			slog.Info("折扣优惠计算拦截，用户不在优惠人群标签范围内", "userId", userID)
			return originalPrice, nil
		}
	}
	return doCalc(originalPrice, discount), nil
}

func minPay(price decimal.Decimal) decimal.Decimal {
	if price.LessThanOrEqual(decimal.Zero) {
		return decimal.NewFromFloat(0.01)
	}
	return price
}

// Registry 折扣策略注册表
type Registry struct {
	m map[string]IDiscountCalculateService
}

func NewRegistry(services ...IDiscountCalculateService) *Registry {
	r := &Registry{m: make(map[string]IDiscountCalculateService)}
	for _, s := range services {
		r.m[s.Plan()] = s
	}
	return r
}

func (r *Registry) Get(plan string) IDiscountCalculateService {
	return r.m[plan]
}

func (r *Registry) Keys() []string {
	keys := make([]string, 0, len(r.m))
	for k := range r.m {
		keys = append(keys, k)
	}
	return keys
}

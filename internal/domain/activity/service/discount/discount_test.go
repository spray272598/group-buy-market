package discount

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"

	"group-buy-market/internal/domain/activity/model/entity"
	"group-buy-market/internal/domain/activity/model/valobj"
)

// mockRepo 领域测试用桩，不依赖基础设施
type mockRepo struct{}

func (m *mockRepo) QueryGroupBuyActivityDiscountVO(ctx context.Context, activityID int64) (*valobj.GroupBuyActivityDiscountVO, error) {
	return nil, nil
}
func (m *mockRepo) QuerySkuByGoodsID(ctx context.Context, goodsID string) (*valobj.SkuVO, error) {
	return nil, nil
}
func (m *mockRepo) QuerySCSkuActivityBySCGoodsID(ctx context.Context, source, channel, goodsID string) (*valobj.SCSkuActivityVO, error) {
	return nil, nil
}
func (m *mockRepo) IsTagCrowdRange(ctx context.Context, tagID, userID string) (bool, error) {
	return true, nil
}
func (m *mockRepo) DowngradeSwitch() bool { return false }
func (m *mockRepo) CutRange(userID string) bool { return true }
func (m *mockRepo) QueryInProgressUserGroupBuyOrderDetailListByOwner(ctx context.Context, activityID int64, userID string, ownerCount int) ([]*entity.UserGroupBuyOrderDetailEntity, error) {
	return nil, nil
}
func (m *mockRepo) QueryInProgressUserGroupBuyOrderDetailListByRandom(ctx context.Context, activityID int64, userID string, randomCount int) ([]*entity.UserGroupBuyOrderDetailEntity, error) {
	return nil, nil
}
func (m *mockRepo) QueryTeamStatisticByActivityID(ctx context.Context, activityID int64) (*valobj.TeamStatisticVO, error) {
	return nil, nil
}

func TestZJ(t *testing.T) {
	svc := NewZJ(&mockRepo{})
	price, err := svc.Calculate(context.Background(), "u1", decimal.NewFromInt(100), &valobj.GroupBuyDiscount{
		DiscountType: valobj.DiscountTypeBase,
		MarketPlan:   "ZJ",
		MarketExpr:   "20",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !price.Equal(decimal.NewFromInt(80)) {
		t.Fatalf("want 80 got %s", price)
	}
}

func TestMJ(t *testing.T) {
	svc := NewMJ(&mockRepo{})
	price, err := svc.Calculate(context.Background(), "u1", decimal.NewFromInt(100), &valobj.GroupBuyDiscount{
		DiscountType: valobj.DiscountTypeBase,
		MarketPlan:   "MJ",
		MarketExpr:   "100,10",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !price.Equal(decimal.NewFromInt(90)) {
		t.Fatalf("want 90 got %s", price)
	}
}

func TestN(t *testing.T) {
	svc := NewN(&mockRepo{})
	price, err := svc.Calculate(context.Background(), "u1", decimal.NewFromInt(100), &valobj.GroupBuyDiscount{
		DiscountType: valobj.DiscountTypeBase,
		MarketPlan:   "N",
		MarketExpr:   "9.9",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !price.Equal(decimal.RequireFromString("9.9")) {
		t.Fatalf("want 9.9 got %s", price)
	}
}

func TestZK(t *testing.T) {
	svc := NewZK(&mockRepo{})
	price, err := svc.Calculate(context.Background(), "u1", decimal.NewFromInt(100), &valobj.GroupBuyDiscount{
		DiscountType: valobj.DiscountTypeBase,
		MarketPlan:   "ZK",
		MarketExpr:   "0.8",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !price.Equal(decimal.NewFromInt(80)) {
		t.Fatalf("want 80 got %s", price)
	}
}

func TestRegistry(t *testing.T) {
	repo := &mockRepo{}
	reg := NewRegistry(NewZJ(repo), NewMJ(repo), NewN(repo), NewZK(repo))
	if reg.Get("ZJ") == nil || reg.Get("MJ") == nil || reg.Get("N") == nil || reg.Get("ZK") == nil {
		t.Fatal("registry incomplete")
	}
	if reg.Get("XX") != nil {
		t.Fatal("unknown plan should be nil")
	}
}

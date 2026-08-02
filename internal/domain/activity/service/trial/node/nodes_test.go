package node

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"group-buy-market/internal/domain/activity/model/entity"
	"group-buy-market/internal/domain/activity/model/valobj"
	"group-buy-market/internal/domain/activity/service/discount"
	"group-buy-market/internal/types/enums"
	"group-buy-market/internal/types/exception"
)

type mockActRepo struct {
	downgrade bool
	cutOK     bool
	activity  *valobj.GroupBuyActivityDiscountVO
	sku       *valobj.SkuVO
	sc        *valobj.SCSkuActivityVO
	tagOK     bool
}

func (m *mockActRepo) QueryGroupBuyActivityDiscountVO(ctx context.Context, activityID int64) (*valobj.GroupBuyActivityDiscountVO, error) {
	return m.activity, nil
}
func (m *mockActRepo) QuerySkuByGoodsID(ctx context.Context, goodsID string) (*valobj.SkuVO, error) {
	return m.sku, nil
}
func (m *mockActRepo) QuerySCSkuActivityBySCGoodsID(ctx context.Context, source, channel, goodsID string) (*valobj.SCSkuActivityVO, error) {
	return m.sc, nil
}
func (m *mockActRepo) IsTagCrowdRange(ctx context.Context, tagID, userID string) (bool, error) {
	return m.tagOK, nil
}
func (m *mockActRepo) DowngradeSwitch() bool { return m.downgrade }
func (m *mockActRepo) CutRange(userID string) bool {
	return m.cutOK
}
func (m *mockActRepo) QueryInProgressUserGroupBuyOrderDetailListByOwner(ctx context.Context, activityID int64, userID string, ownerCount int) ([]*entity.UserGroupBuyOrderDetailEntity, error) {
	return nil, nil
}
func (m *mockActRepo) QueryInProgressUserGroupBuyOrderDetailListByRandom(ctx context.Context, activityID int64, userID string, randomCount int) ([]*entity.UserGroupBuyOrderDetailEntity, error) {
	return nil, nil
}
func (m *mockActRepo) QueryTeamStatisticByActivityID(ctx context.Context, activityID int64) (*valobj.TeamStatisticVO, error) {
	return &valobj.TeamStatisticVO{}, nil
}

func sampleActivity() *valobj.GroupBuyActivityDiscountVO {
	return &valobj.GroupBuyActivityDiscountVO{
		ActivityID:   100123,
		ActivityName: "测试",
		GroupBuyDiscount: &valobj.GroupBuyDiscount{
			DiscountName: "直减",
			DiscountType: valobj.DiscountTypeBase,
			MarketPlan:   "ZJ",
			MarketExpr:   "20",
		},
		Target:    3,
		ValidTime: 15,
		StartTime: time.Now().Add(-time.Hour),
		EndTime:   time.Now().Add(24 * time.Hour),
		TagID:     "",
		TagScope:  "",
	}
}

func TestTrialSuccess(t *testing.T) {
	repo := &mockActRepo{
		cutOK:    true,
		activity: sampleActivity(),
		sku: &valobj.SkuVO{
			GoodsID:       "9890001",
			GoodsName:     "书",
			OriginalPrice: decimal.NewFromInt(100),
		},
		sc: &valobj.SCSkuActivityVO{ActivityID: 100123, GoodsID: "9890001", Source: "s01", Channel: "c01"},
	}
	reg := discount.NewRegistry(discount.NewZJ(repo), discount.NewMJ(repo), discount.NewN(repo), discount.NewZK(repo))
	chain := NewChain(repo, reg)
	res, err := chain.Apply(context.Background(), &entity.MarketProductEntity{
		UserID: "xfg01", GoodsID: "9890001", Source: "s01", Channel: "c01",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.PayPrice.Equal(decimal.NewFromInt(80)) {
		t.Fatalf("pay want 80 got %s", res.PayPrice)
	}
	if !res.IsVisible || !res.IsEnable {
		t.Fatal("should visible enable")
	}
}

func TestTrialIllegalParam(t *testing.T) {
	repo := &mockActRepo{cutOK: true}
	reg := discount.NewRegistry(discount.NewZJ(repo))
	chain := NewChain(repo, reg)
	_, err := chain.Apply(context.Background(), &entity.MarketProductEntity{})
	ae, ok := exception.AsAppException(err)
	if !ok || ae.Code != enums.ILLEGAL_PARAMETER.Code {
		t.Fatalf("want illegal param, got %v", err)
	}
}

func TestTrialDowngrade(t *testing.T) {
	repo := &mockActRepo{downgrade: true, cutOK: true}
	reg := discount.NewRegistry(discount.NewZJ(repo))
	chain := NewChain(repo, reg)
	_, err := chain.Apply(context.Background(), &entity.MarketProductEntity{
		UserID: "u", GoodsID: "g", Source: "s", Channel: "c",
	})
	ae, ok := exception.AsAppException(err)
	if !ok || ae.Code != enums.E0003.Code {
		t.Fatalf("want E0003 got %v", err)
	}
}

func TestTrialNoConfig(t *testing.T) {
	repo := &mockActRepo{cutOK: true, activity: nil, sku: nil, sc: nil}
	reg := discount.NewRegistry(discount.NewZJ(repo))
	chain := NewChain(repo, reg)
	_, err := chain.Apply(context.Background(), &entity.MarketProductEntity{
		UserID: "u", GoodsID: "g", Source: "s", Channel: "c",
	})
	ae, ok := exception.AsAppException(err)
	if !ok || ae.Code != enums.E0002.Code {
		t.Fatalf("want E0002 got %v", err)
	}
}

func TestTrialTagLimit(t *testing.T) {
	act := sampleActivity()
	act.TagID = "TAG1"
	act.TagScope = "1,2" // 可见+参与都限制
	repo := &mockActRepo{
		cutOK:    true,
		activity: act,
		tagOK:    false,
		sku: &valobj.SkuVO{
			GoodsID: "9890001", GoodsName: "书", OriginalPrice: decimal.NewFromInt(100),
		},
	}
	id := int64(100123)
	reg := discount.NewRegistry(discount.NewZJ(repo))
	chain := NewChain(repo, reg)
	res, err := chain.Apply(context.Background(), &entity.MarketProductEntity{
		UserID: "outsider", GoodsID: "9890001", Source: "s01", Channel: "c01", ActivityID: &id,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsVisible || res.IsEnable {
		t.Fatal("outsider should not visible/enable")
	}
}

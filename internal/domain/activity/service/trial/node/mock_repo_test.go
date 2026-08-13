package node

import (
	"context"
	"time"

	"github.com/shopspring/decimal"

	"group-buy-market/internal/domain/activity/model/entity"
	"group-buy-market/internal/domain/activity/model/valobj"
	"group-buy-market/internal/domain/activity/service/trial/factory"
)

// mockRepo 模拟真实 DB 网络 RT，用于 benchmark 对比并行/串行收益。
// dbRT 模拟单次 SQL 查询的网络+磁盘往返耗时（本地 Docker MySQL 约 2~8ms）。
type mockRepo struct {
	dbRT time.Duration
}

func (m *mockRepo) query(ctx context.Context) error {
	select {
	case <-time.After(m.dbRT):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (m *mockRepo) QueryGroupBuyActivityDiscountVO(ctx context.Context, activityID int64) (*valobj.GroupBuyActivityDiscountVO, error) {
	if err := m.query(ctx); err != nil {
		return nil, err
	}
	return &valobj.GroupBuyActivityDiscountVO{
		GroupBuyDiscount: &valobj.GroupBuyDiscount{MarketPlan: "ZJ"},
	}, nil
}

func (m *mockRepo) QuerySkuByGoodsID(ctx context.Context, goodsID string) (*valobj.SkuVO, error) {
	if err := m.query(ctx); err != nil {
		return nil, err
	}
	return &valobj.SkuVO{OriginalPrice: decimal.NewFromInt(100)}, nil
}

func (m *mockRepo) QuerySCSkuActivityBySCGoodsID(ctx context.Context, source, channel, goodsID string) (*valobj.SCSkuActivityVO, error) {
	if err := m.query(ctx); err != nil {
		return nil, err
	}
	return &valobj.SCSkuActivityVO{ActivityID: 1}, nil
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

// serialLoad 模拟 Java/旧版串行加载：先 loadActivity(2次查询)，再 QuerySku(1次查询)
func serialLoad(ctx context.Context, c *Chain, req *entity.MarketProductEntity, dc *factory.DynamicContext) error {
	act, err := c.loadActivity(ctx, req)
	if err != nil {
		return err
	}
	sku, err := c.repo.QuerySkuByGoodsID(ctx, req.GoodsID)
	if err != nil {
		return err
	}
	dc.GroupBuyActivityDiscountVO = act
	dc.SkuVO = sku
	return nil
}

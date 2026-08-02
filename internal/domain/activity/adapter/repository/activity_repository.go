package repository

import (
	"context"

	"group-buy-market/internal/domain/activity/model/entity"
	"group-buy-market/internal/domain/activity/model/valobj"
)

// IActivityRepository 活动仓储端口（领域层定义，基础设施实现）
type IActivityRepository interface {
	QueryGroupBuyActivityDiscountVO(ctx context.Context, activityID int64) (*valobj.GroupBuyActivityDiscountVO, error)
	QuerySkuByGoodsID(ctx context.Context, goodsID string) (*valobj.SkuVO, error)
	QuerySCSkuActivityBySCGoodsID(ctx context.Context, source, channel, goodsID string) (*valobj.SCSkuActivityVO, error)
	IsTagCrowdRange(ctx context.Context, tagID, userID string) (bool, error)
	DowngradeSwitch() bool
	CutRange(userID string) bool
	QueryInProgressUserGroupBuyOrderDetailListByOwner(ctx context.Context, activityID int64, userID string, ownerCount int) ([]*entity.UserGroupBuyOrderDetailEntity, error)
	QueryInProgressUserGroupBuyOrderDetailListByRandom(ctx context.Context, activityID int64, userID string, randomCount int) ([]*entity.UserGroupBuyOrderDetailEntity, error)
	QueryTeamStatisticByActivityID(ctx context.Context, activityID int64) (*valobj.TeamStatisticVO, error)
}

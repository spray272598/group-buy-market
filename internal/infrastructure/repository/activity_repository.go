package repository

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"gorm.io/gorm"

	activityrepo "group-buy-market/internal/domain/activity/adapter/repository"
	"group-buy-market/internal/domain/activity/model/entity"
	"group-buy-market/internal/domain/activity/model/valobj"
	"group-buy-market/internal/infrastructure/dao/po"
	"group-buy-market/internal/infrastructure/dcc"
	redisx "group-buy-market/internal/infrastructure/redis"
)

// 编译期校验：基础设施实现领域端口
var _ activityrepo.IActivityRepository = (*ActivityRepository)(nil)

// ActivityRepository 活动仓储实现（Infrastructure 层）
type ActivityRepository struct {
	db    *gorm.DB
	redis *redisx.Service
	dcc   *dcc.Service
}

func NewActivityRepository(db *gorm.DB, rdb *redisx.Service, dccSvc *dcc.Service) *ActivityRepository {
	return &ActivityRepository{db: db, redis: rdb, dcc: dccSvc}
}

func (r *ActivityRepository) QueryGroupBuyActivityDiscountVO(ctx context.Context, activityID int64) (*valobj.GroupBuyActivityDiscountVO, error) {
	var act po.GroupBuyActivity
	err := r.db.WithContext(ctx).
		Where("activity_id = ? AND status = 1", activityID).
		First(&act).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var discount po.GroupBuyDiscount
	err = r.db.WithContext(ctx).
		Where("discount_id = ?", act.DiscountID).
		First(&discount).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &valobj.GroupBuyActivityDiscountVO{
		ActivityID:   act.ActivityID,
		ActivityName: act.ActivityName,
		GroupBuyDiscount: &valobj.GroupBuyDiscount{
			DiscountName: discount.DiscountName,
			DiscountDesc: discount.DiscountDesc,
			DiscountType: valobj.DiscountType(discount.DiscountType),
			MarketPlan:   discount.MarketPlan,
			MarketExpr:   discount.MarketExpr,
			TagID:        discount.TagID,
		},
		GroupType:      act.GroupType,
		TakeLimitCount: act.TakeLimitCount,
		Target:         act.Target,
		ValidTime:      act.ValidTime,
		Status:         act.Status,
		StartTime:      act.StartTime,
		EndTime:        act.EndTime,
		TagID:          act.TagID,
		TagScope:       act.TagScope,
	}, nil
}

func (r *ActivityRepository) QuerySkuByGoodsID(ctx context.Context, goodsID string) (*valobj.SkuVO, error) {
	var sku po.Sku
	err := r.db.WithContext(ctx).Where("goods_id = ?", goodsID).First(&sku).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &valobj.SkuVO{
		GoodsID:       sku.GoodsID,
		GoodsName:     sku.GoodsName,
		OriginalPrice: sku.OriginalPrice,
	}, nil
}

func (r *ActivityRepository) QuerySCSkuActivityBySCGoodsID(ctx context.Context, source, channel, goodsID string) (*valobj.SCSkuActivityVO, error) {
	var row po.SCSkuActivity
	err := r.db.WithContext(ctx).
		Where("source = ? AND channel = ? AND goods_id = ?", source, channel, goodsID).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &valobj.SCSkuActivityVO{
		Source:     row.Source,
		Channel:    row.Channel,
		ActivityID: row.ActivityID,
		GoodsID:    row.GoodsID,
	}, nil
}

func (r *ActivityRepository) IsTagCrowdRange(ctx context.Context, tagID, userID string) (bool, error) {
	// 优先 Redis BitSet；无 key 时回退 DB 明细
	ok, err := r.redis.IsTagCrowdRange(ctx, tagID, userID)
	if err != nil {
		return false, err
	}
	// BitSet 不存在时 Java 返回 true；这里保持一致
	// 若存在 BitSet 则用其结果
	exists, err := r.redis.Client().Exists(ctx, tagID).Result()
	if err != nil {
		return false, err
	}
	if exists > 0 {
		return ok, nil
	}
	// 回退：查 crowd_tags_detail
	var count int64
	err = r.db.WithContext(ctx).Model(&po.CrowdTagsDetail{}).
		Where("tag_id = ? AND user_id = ?", tagID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	// 无明细配置时放行（与 bitset 不存在一致）
	var total int64
	_ = r.db.WithContext(ctx).Model(&po.CrowdTagsDetail{}).Where("tag_id = ?", tagID).Count(&total).Error
	if total == 0 {
		return true, nil
	}
	return count > 0, nil
}

func (r *ActivityRepository) DowngradeSwitch() bool {
	return r.dcc.IsDowngradeSwitch()
}

func (r *ActivityRepository) CutRange(userID string) bool {
	return r.dcc.IsCutRange(userID)
}

func (r *ActivityRepository) QueryInProgressUserGroupBuyOrderDetailListByOwner(ctx context.Context, activityID int64, userID string, ownerCount int) ([]*entity.UserGroupBuyOrderDetailEntity, error) {
	var lists []po.GroupBuyOrderList
	err := r.db.WithContext(ctx).
		Select("user_id, team_id, out_trade_no").
		Where("activity_id = ? AND user_id = ? AND status IN (0,1) AND end_time > ?", activityID, userID, time.Now()).
		Order("id DESC").
		Limit(ownerCount).
		Find(&lists).Error
	if err != nil {
		return nil, err
	}
	return r.assembleOrderDetails(ctx, lists)
}

func (r *ActivityRepository) QueryInProgressUserGroupBuyOrderDetailListByRandom(ctx context.Context, activityID int64, userID string, randomCount int) ([]*entity.UserGroupBuyOrderDetailEntity, error) {
	var lists []po.GroupBuyOrderList
	err := r.db.WithContext(ctx).
		Select("user_id, team_id, out_trade_no").
		Where(`activity_id = ? AND user_id != ? AND status IN (0,1) AND end_time > ?
			AND team_id IN (SELECT team_id FROM group_buy_order WHERE activity_id = ? AND status = 0)`,
			activityID, userID, time.Now(), activityID).
		Order("id DESC").
		Limit(randomCount * 2).
		Find(&lists).Error
	if err != nil {
		return nil, err
	}
	if len(lists) > randomCount {
		rand.Shuffle(len(lists), func(i, j int) { lists[i], lists[j] = lists[j], lists[i] })
		lists = lists[:randomCount]
	}
	return r.assembleOrderDetails(ctx, lists)
}

func (r *ActivityRepository) assembleOrderDetails(ctx context.Context, lists []po.GroupBuyOrderList) ([]*entity.UserGroupBuyOrderDetailEntity, error) {
	if len(lists) == 0 {
		return nil, nil
	}
	teamIDs := make([]string, 0, len(lists))
	seen := map[string]struct{}{}
	for _, l := range lists {
		if l.TeamID == "" {
			continue
		}
		if _, ok := seen[l.TeamID]; !ok {
			seen[l.TeamID] = struct{}{}
			teamIDs = append(teamIDs, l.TeamID)
		}
	}
	var orders []po.GroupBuyOrder
	err := r.db.WithContext(ctx).
		Where("status = 0 AND target_count > lock_count AND valid_end_time > ? AND team_id IN ?", time.Now(), teamIDs).
		Find(&orders).Error
	if err != nil {
		return nil, err
	}
	orderMap := make(map[string]po.GroupBuyOrder, len(orders))
	for _, o := range orders {
		orderMap[o.TeamID] = o
	}
	result := make([]*entity.UserGroupBuyOrderDetailEntity, 0, len(lists))
	for _, l := range lists {
		o, ok := orderMap[l.TeamID]
		if !ok {
			continue
		}
		result = append(result, &entity.UserGroupBuyOrderDetailEntity{
			UserID:         l.UserID,
			TeamID:         o.TeamID,
			ActivityID:     o.ActivityID,
			TargetCount:    o.TargetCount,
			CompleteCount:  o.CompleteCount,
			LockCount:      o.LockCount,
			ValidStartTime: o.ValidStartTime,
			ValidEndTime:   o.ValidEndTime,
			OutTradeNo:     l.OutTradeNo,
		})
	}
	return result, nil
}

func (r *ActivityRepository) QueryTeamStatisticByActivityID(ctx context.Context, activityID int64) (*valobj.TeamStatisticVO, error) {
	var teamIDs []string
	err := r.db.WithContext(ctx).Model(&po.GroupBuyOrderList{}).
		Select("DISTINCT team_id").
		Where("activity_id = ? AND status IN (0,1)", activityID).
		Pluck("team_id", &teamIDs).Error
	if err != nil {
		return nil, err
	}
	if len(teamIDs) == 0 {
		return &valobj.TeamStatisticVO{}, nil
	}
	var allTeam int64
	var completeTeam int64
	var userCount int64
	_ = r.db.WithContext(ctx).Model(&po.GroupBuyOrder{}).Where("team_id IN ?", teamIDs).Count(&allTeam).Error
	_ = r.db.WithContext(ctx).Model(&po.GroupBuyOrder{}).Where("status = 1 AND team_id IN ?", teamIDs).Count(&completeTeam).Error
	r.db.WithContext(ctx).Model(&po.GroupBuyOrder{}).Select("COALESCE(SUM(lock_count),0)").Where("team_id IN ?", teamIDs).Scan(&userCount)
	return &valobj.TeamStatisticVO{
		AllTeamCount:         allTeam,
		AllTeamCompleteCount: completeTeam,
		AllTeamUserCount:     userCount,
	}, nil
}

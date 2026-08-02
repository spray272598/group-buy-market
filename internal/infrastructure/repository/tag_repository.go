package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	tagrepo "group-buy-market/internal/domain/tag/adapter/repository"
	"group-buy-market/internal/domain/tag/model/entity"
	"group-buy-market/internal/infrastructure/dao/po"
	redisx "group-buy-market/internal/infrastructure/redis"
)

var _ tagrepo.ITagRepository = (*TagRepository)(nil)

// TagRepository 人群标签仓储：DB 明细 + Redis BitSet
type TagRepository struct {
	db    *gorm.DB
	redis *redisx.Service
}

func NewTagRepository(db *gorm.DB, rdb *redisx.Service) *TagRepository {
	return &TagRepository{db: db, redis: rdb}
}

func (r *TagRepository) QueryCrowdTagsJobEntity(ctx context.Context, tagID, batchID string) (*entity.CrowdTagsJobEntity, error) {
	var job po.CrowdTagsJob
	err := r.db.WithContext(ctx).Where("tag_id = ? AND batch_id = ?", tagID, batchID).First(&job).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entity.CrowdTagsJobEntity{
		TagType:       job.TagType,
		TagRule:       job.TagRule,
		StatStartTime: job.StatStartTime,
		StatEndTime:   job.StatEndTime,
	}, nil
}

func (r *TagRepository) AddCrowdTagsUserID(ctx context.Context, tagID, userID string) error {
	now := time.Now()
	detail := po.CrowdTagsDetail{
		TagID:      tagID,
		UserID:     userID,
		CreateTime: now,
		UpdateTime: now,
	}
	// 唯一索引冲突忽略
	if err := r.db.WithContext(ctx).Where("tag_id = ? AND user_id = ?", tagID, userID).
		FirstOrCreate(&detail).Error; err != nil {
		return err
	}
	// Redis BitSet 打标（试算 TagNode 使用）
	return r.redis.AddTagCrowd(ctx, tagID, userID)
}

func (r *TagRepository) UpdateCrowdTagsStatistics(ctx context.Context, tagID string, count int) error {
	return r.db.WithContext(ctx).Model(&po.CrowdTags{}).
		Where("tag_id = ?", tagID).
		Updates(map[string]any{
			"statistics":  gorm.Expr("statistics + ?", count),
			"update_time": time.Now(),
		}).Error
}

package repository

import (
	"context"

	"group-buy-market/internal/domain/tag/model/entity"
)

// ITagRepository 人群标签仓储端口
type ITagRepository interface {
	QueryCrowdTagsJobEntity(ctx context.Context, tagID, batchID string) (*entity.CrowdTagsJobEntity, error)
	AddCrowdTagsUserID(ctx context.Context, tagID, userID string) error
	UpdateCrowdTagsStatistics(ctx context.Context, tagID string, count int) error
}

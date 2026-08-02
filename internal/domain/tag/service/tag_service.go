package service

import (
	"context"
	"log/slog"

	"group-buy-market/internal/domain/tag/adapter/repository"
)

// ITagService 人群标签领域服务
type ITagService interface {
	ExecTagBatchJob(ctx context.Context, tagID, batchID string) error
}

type TagService struct {
	repo repository.ITagRepository
}

func NewTagService(repo repository.ITagRepository) *TagService {
	return &TagService{repo: repo}
}

// ExecTagBatchJob 执行人群标签批次任务（对齐 Java TagService）
// 生产中通常由数仓写入，这里模拟采集用户并写入 DB + Redis BitSet
func (s *TagService) ExecTagBatchJob(ctx context.Context, tagID, batchID string) error {
	slog.Info("人群标签批次任务", "tagId", tagID, "batchId", batchID)

	job, err := s.repo.QueryCrowdTagsJobEntity(ctx, tagID, batchID)
	if err != nil {
		return err
	}
	if job == nil {
		slog.Warn("人群标签批次任务不存在", "tagId", tagID, "batchId", batchID)
		return nil
	}

	// 模拟采集用户（原项目写死示例用户，后续可对接订单数据）
	userIDs := []string{
		"xiaofuge", "liergou",
		"xfg01", "xfg02", "xfg03", "xfg04", "xfg05",
		"xfg06", "xfg07", "xfg08", "xfg09",
	}
	for _, uid := range userIDs {
		if err := s.repo.AddCrowdTagsUserID(ctx, tagID, uid); err != nil {
			return err
		}
	}
	return s.repo.UpdateCrowdTagsStatistics(ctx, tagID, len(userIDs))
}

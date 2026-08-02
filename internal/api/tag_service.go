package api

import (
	"context"

	"group-buy-market/internal/api/dto"
	"group-buy-market/internal/api/response"
)

// ITagService 人群标签对外契约
type ITagService interface {
	ExecTagBatchJob(ctx context.Context, req *dto.TagBatchJobRequestDTO) response.Response[bool]
}

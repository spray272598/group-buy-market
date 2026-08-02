package http

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"group-buy-market/internal/api"
	"group-buy-market/internal/api/dto"
	"group-buy-market/internal/api/response"
	"group-buy-market/internal/domain/tag/service"
	"group-buy-market/internal/types/enums"
)

var _ api.ITagService = (*TagController)(nil)

// TagController 人群标签 Trigger（实现 api.ITagService）
type TagController struct {
	tagService service.ITagService
}

func NewTagController(tagService service.ITagService) *TagController {
	return &TagController{tagService: tagService}
}

func (c *TagController) Register(r *gin.Engine) {
	g := r.Group("/api/v1/gbm/tag")
	handler := func(ctx *gin.Context) {
		tagID := ctx.Query("tagId")
		batchID := ctx.Query("batchId")
		if tagID == "" || batchID == "" {
			var body dto.TagBatchJobRequestDTO
			_ = ctx.ShouldBindJSON(&body)
			if tagID == "" {
				tagID = body.TagID
			}
			if batchID == "" {
				batchID = body.BatchID
			}
		}
		ctx.JSON(http.StatusOK, c.ExecTagBatchJob(ctx.Request.Context(), &dto.TagBatchJobRequestDTO{
			TagID: tagID, BatchID: batchID,
		}))
	}
	g.GET("/exec_tag_batch_job", handler)
	g.POST("/exec_tag_batch_job", handler)
}

func (c *TagController) ExecTagBatchJob(ctx context.Context, req *dto.TagBatchJobRequestDTO) response.Response[bool] {
	if req == nil || req.TagID == "" || req.BatchID == "" {
		return response.Fail[bool](enums.ILLEGAL_PARAMETER.Code, enums.ILLEGAL_PARAMETER.Info)
	}
	slog.Info("执行人群标签批次", "tagId", req.TagID, "batchId", req.BatchID)
	if err := c.tagService.ExecTagBatchJob(ctx, req.TagID, req.BatchID); err != nil {
		slog.Error("人群标签批次失败", "err", err)
		return response.Fail[bool](enums.UN_ERROR.Code, enums.UN_ERROR.Info)
	}
	return response.Success(true)
}

package http

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"group-buy-market/internal/domain/tag/service"
	"group-buy-market/internal/types/enums"
	"group-buy-market/internal/types/response"
)

// TagController 人群标签 Trigger
type TagController struct {
	tagService service.ITagService
}

func NewTagController(tagService service.ITagService) *TagController {
	return &TagController{tagService: tagService}
}

func (c *TagController) Register(r *gin.Engine) {
	g := r.Group("/api/v1/gbm/tag")
	g.GET("/exec_tag_batch_job", c.ExecTagBatchJob)
	g.POST("/exec_tag_batch_job", c.ExecTagBatchJob)
}

// ExecTagBatchJob 执行人群打标批次
// GET /api/v1/gbm/tag/exec_tag_batch_job?tagId=xxx&batchId=10001
func (c *TagController) ExecTagBatchJob(ctx *gin.Context) {
	tagID := ctx.Query("tagId")
	batchID := ctx.Query("batchId")
	if tagID == "" || batchID == "" {
		var body struct {
			TagID   string `json:"tagId"`
			BatchID string `json:"batchId"`
		}
		_ = ctx.ShouldBindJSON(&body)
		if tagID == "" {
			tagID = body.TagID
		}
		if batchID == "" {
			batchID = body.BatchID
		}
	}
	if tagID == "" || batchID == "" {
		ctx.JSON(http.StatusOK, response.Fail[bool](enums.ILLEGAL_PARAMETER.Code, enums.ILLEGAL_PARAMETER.Info))
		return
	}
	slog.Info("执行人群标签批次", "tagId", tagID, "batchId", batchID)
	if err := c.tagService.ExecTagBatchJob(ctx.Request.Context(), tagID, batchID); err != nil {
		slog.Error("人群标签批次失败", "err", err)
		ctx.JSON(http.StatusOK, response.Fail[bool](enums.UN_ERROR.Code, enums.UN_ERROR.Info))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(true))
}

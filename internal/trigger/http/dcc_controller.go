package http

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"group-buy-market/internal/api"
	"group-buy-market/internal/api/dto"
	"group-buy-market/internal/api/response"
	"group-buy-market/internal/infrastructure/dcc"
	"group-buy-market/internal/types/enums"
)

var _ api.IDCCService = (*DCCController)(nil)

// DCCController 动态配置 Trigger（实现 api.IDCCService）
type DCCController struct {
	dcc *dcc.Service
}

func NewDCCController(svc *dcc.Service) *DCCController {
	return &DCCController{dcc: svc}
}

func (c *DCCController) Register(r *gin.Engine) {
	g := r.Group("/api/v1/gbm/dcc")
	g.GET("/query", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, c.Query(ctx.Request.Context()))
	})
	g.GET("/update_config", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, c.UpdateConfig(ctx.Request.Context(), ctx.Query("key"), ctx.Query("value")))
	})
	g.POST("/update", func(ctx *gin.Context) {
		var req dto.DCCUpdateRequestDTO
		if err := ctx.ShouldBindJSON(&req); err != nil || req.Key == "" {
			ctx.JSON(http.StatusOK, response.Fail[any](enums.ILLEGAL_PARAMETER.Code, enums.ILLEGAL_PARAMETER.Info))
			return
		}
		ctx.JSON(http.StatusOK, c.UpdateConfig(ctx.Request.Context(), req.Key, req.Value))
	})
}

func (c *DCCController) Query(ctx context.Context) response.Response[dto.DCCSnapshot] {
	return response.Success(dto.DCCSnapshot(c.dcc.Snapshot()))
}

func (c *DCCController) UpdateConfig(ctx context.Context, key, value string) response.Response[bool] {
	if key == "" {
		return response.Fail[bool](enums.ILLEGAL_PARAMETER.Code, enums.ILLEGAL_PARAMETER.Info)
	}
	slog.Info("DCC 动态配置值变更", "key", key, "value", value)
	c.dcc.Update(key, value)
	return response.Success(true)
}

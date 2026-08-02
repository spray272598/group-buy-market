package http

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"group-buy-market/internal/api"
	"group-buy-market/internal/api/dto"
	"group-buy-market/internal/api/response"
	"group-buy-market/internal/types/enums"
)

// MarketIndexController 仅做 HTTP 协议适配 + 限流；用例在 application
type MarketIndexController struct {
	app       api.IMarketIndexService
	rateLimit *RateLimitStore
}

func NewMarketIndexController(app api.IMarketIndexService, rl *RateLimitStore) *MarketIndexController {
	return &MarketIndexController{app: app, rateLimit: rl}
}

func (c *MarketIndexController) Register(r *gin.Engine) {
	g := r.Group("/api/v1/gbm/index")
	g.POST("/query_group_buy_market_config", c.handleQuery)
}

func (c *MarketIndexController) handleQuery(ctx *gin.Context) {
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail[dto.GoodsMarketResponseDTO](enums.ILLEGAL_PARAMETER.Code, enums.ILLEGAL_PARAMETER.Info))
		return
	}
	var req dto.GoodsMarketRequestDTO
	if err := json.Unmarshal(body, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail[dto.GoodsMarketResponseDTO](enums.ILLEGAL_PARAMETER.Code, enums.ILLEGAL_PARAMETER.Info))
		return
	}
	if c.rateLimit != nil && !c.rateLimit.allow(req.UserID) {
		slog.Error("查询拼团营销配置限流", "userId", req.UserID)
		ctx.JSON(http.StatusOK, response.Fail[dto.GoodsMarketResponseDTO](enums.RATE_LIMITER.Code, enums.RATE_LIMITER.Info))
		return
	}
	ctx.JSON(http.StatusOK, c.app.QueryGroupBuyMarketConfig(ctx.Request.Context(), &req))
}

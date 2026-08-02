package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"group-buy-market/internal/api"
	"group-buy-market/internal/api/dto"
	"group-buy-market/internal/api/response"
	"group-buy-market/internal/types/enums"
)

// MarketTradeController 仅 HTTP 适配；用例在 application
type MarketTradeController struct {
	app api.IMarketTradeService
}

func NewMarketTradeController(app api.IMarketTradeService) *MarketTradeController {
	return &MarketTradeController{app: app}
}

func (c *MarketTradeController) Register(r *gin.Engine) {
	g := r.Group("/api/v1/gbm/trade")
	g.POST("/lock_market_pay_order", c.handleLock)
	g.POST("/settlement_market_pay_order", c.handleSettlement)
	g.POST("/refund_market_pay_order", c.handleRefund)
}

func (c *MarketTradeController) handleLock(ctx *gin.Context) {
	var req dto.LockMarketPayOrderRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail[dto.LockMarketPayOrderResponseDTO](enums.ILLEGAL_PARAMETER.Code, enums.ILLEGAL_PARAMETER.Info))
		return
	}
	ctx.JSON(http.StatusOK, c.app.LockMarketPayOrder(ctx.Request.Context(), &req))
}

func (c *MarketTradeController) handleSettlement(ctx *gin.Context) {
	var req dto.SettlementMarketPayOrderRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail[dto.SettlementMarketPayOrderResponseDTO](enums.ILLEGAL_PARAMETER.Code, enums.ILLEGAL_PARAMETER.Info))
		return
	}
	ctx.JSON(http.StatusOK, c.app.SettlementMarketPayOrder(ctx.Request.Context(), &req))
}

func (c *MarketTradeController) handleRefund(ctx *gin.Context) {
	var req dto.RefundMarketPayOrderRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail[dto.RefundMarketPayOrderResponseDTO](enums.ILLEGAL_PARAMETER.Code, enums.ILLEGAL_PARAMETER.Info))
		return
	}
	ctx.JSON(http.StatusOK, c.app.RefundMarketPayOrder(ctx.Request.Context(), &req))
}

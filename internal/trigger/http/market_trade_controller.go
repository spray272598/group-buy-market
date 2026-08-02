package http

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	activityentity "group-buy-market/internal/domain/activity/model/entity"
	activityservice "group-buy-market/internal/domain/activity/service"
	"group-buy-market/internal/domain/trade/model/entity"
	"group-buy-market/internal/domain/trade/model/valobj"
	"group-buy-market/internal/domain/trade/service/lock"
	"group-buy-market/internal/domain/trade/service/refund"
	"group-buy-market/internal/domain/trade/service/settlement"
	"group-buy-market/internal/types/enums"
	"group-buy-market/internal/types/exception"
	"group-buy-market/internal/types/response"
)

// MarketTradeController 营销交易 Trigger
type MarketTradeController struct {
	indexService      activityservice.IIndexGroupBuyMarketService
	lockService       lock.ITradeLockOrderService
	settlementService settlement.ITradeSettlementOrderService
	refundService     refund.ITradeRefundOrderService
}

func NewMarketTradeController(
	indexService activityservice.IIndexGroupBuyMarketService,
	lockService lock.ITradeLockOrderService,
	settlementService settlement.ITradeSettlementOrderService,
	refundService refund.ITradeRefundOrderService,
) *MarketTradeController {
	return &MarketTradeController{
		indexService:      indexService,
		lockService:       lockService,
		settlementService: settlementService,
		refundService:     refundService,
	}
}

func (c *MarketTradeController) Register(r *gin.Engine) {
	g := r.Group("/api/v1/gbm/trade")
	g.POST("/lock_market_pay_order", c.LockMarketPayOrder)
	g.POST("/settlement_market_pay_order", c.SettlementMarketPayOrder)
	g.POST("/refund_market_pay_order", c.RefundMarketPayOrder)
}

func (c *MarketTradeController) LockMarketPayOrder(ctx *gin.Context) {
	var req LockMarketPayOrderRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail[LockMarketPayOrderResponseDTO](enums.ILLEGAL_PARAMETER.Code, enums.ILLEGAL_PARAMETER.Info))
		return
	}

	// 兼容 notifyUrl 字段
	if req.NotifyConfig == nil && req.NotifyURL != "" {
		req.NotifyConfig = &NotifyConfigDTO{NotifyType: "HTTP", NotifyUrl: req.NotifyURL}
	}
	if req.NotifyConfig == nil {
		req.NotifyConfig = &NotifyConfigDTO{NotifyType: "MQ"}
	}

	if req.UserID == "" || req.Source == "" || req.Channel == "" || req.GoodsID == "" || req.ActivityID == 0 ||
		(req.NotifyConfig.NotifyType == "HTTP" && req.NotifyConfig.NotifyUrl == "") {
		ctx.JSON(http.StatusOK, response.Fail[LockMarketPayOrderResponseDTO](enums.ILLEGAL_PARAMETER.Code, enums.ILLEGAL_PARAMETER.Info))
		return
	}

	slog.Info("营销交易锁单", "userId", req.UserID, "outTradeNo", req.OutTradeNo)

	// 幂等：已存在 CREATE 状态锁单
	exist, err := c.lockService.QueryNoPayMarketPayOrderByOutTradeNo(ctx.Request.Context(), req.UserID, req.OutTradeNo)
	if err != nil {
		c.failLock(ctx, err)
		return
	}
	if exist != nil && exist.TradeOrderStatus == valobj.TradeOrderCreate {
		ctx.JSON(http.StatusOK, response.Success(LockMarketPayOrderResponseDTO{
			OrderID:          exist.OrderID,
			OriginalPrice:    exist.OriginalPrice,
			DeductionPrice:   exist.DeductionPrice,
			PayPrice:         exist.PayPrice,
			TradeOrderStatus: int(exist.TradeOrderStatus),
			TeamID:           exist.TeamID,
		}))
		return
	}

	// 拼团目标已达成拦截
	if req.TeamID != "" {
		progress, err := c.lockService.QueryGroupBuyProgress(ctx.Request.Context(), req.TeamID)
		if err != nil {
			c.failLock(ctx, err)
			return
		}
		if progress != nil && progress.TargetCount == progress.LockCount {
			ctx.JSON(http.StatusOK, response.Fail[LockMarketPayOrderResponseDTO](enums.E0006.Code, enums.E0006.Info))
			return
		}
	}

	// 试算
	activityID := req.ActivityID
	trial, err := c.indexService.IndexMarketTrial(ctx.Request.Context(), &activityentity.MarketProductEntity{
		UserID:     req.UserID,
		Source:     req.Source,
		Channel:    req.Channel,
		GoodsID:    req.GoodsID,
		ActivityID: &activityID,
	})
	if err != nil {
		c.failLock(ctx, err)
		return
	}
	if !trial.IsVisible || !trial.IsEnable {
		ctx.JSON(http.StatusOK, response.Fail[LockMarketPayOrderResponseDTO](enums.E0007.Code, enums.E0007.Info))
		return
	}

	g := trial.GroupBuyActivityDiscountVO
	order, err := c.lockService.LockMarketPayOrder(
		ctx.Request.Context(),
		&entity.UserEntity{UserID: req.UserID},
		&entity.PayActivityEntity{
			TeamID:       req.TeamID,
			ActivityID:   req.ActivityID,
			ActivityName: g.ActivityName,
			StartTime:    g.StartTime,
			EndTime:      g.EndTime,
			ValidTime:    g.ValidTime,
			TargetCount:  g.Target,
		},
		&entity.PayDiscountEntity{
			Source:         req.Source,
			Channel:        req.Channel,
			GoodsID:        req.GoodsID,
			GoodsName:      trial.GoodsName,
			OriginalPrice:  trial.OriginalPrice,
			DeductionPrice: trial.DeductionPrice,
			PayPrice:       trial.PayPrice,
			OutTradeNo:     req.OutTradeNo,
			NotifyConfig: &valobj.NotifyConfigVO{
				NotifyType: valobj.NotifyType(req.NotifyConfig.NotifyType),
				NotifyMQ:   req.NotifyConfig.NotifyMQ,
				NotifyUrl:  req.NotifyConfig.NotifyUrl,
			},
		},
	)
	if err != nil {
		c.failLock(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response.Success(LockMarketPayOrderResponseDTO{
		OrderID:          order.OrderID,
		OriginalPrice:    order.OriginalPrice,
		DeductionPrice:   order.DeductionPrice,
		PayPrice:         order.PayPrice,
		TradeOrderStatus: int(order.TradeOrderStatus),
		TeamID:           order.TeamID,
	}))
}

func (c *MarketTradeController) SettlementMarketPayOrder(ctx *gin.Context) {
	var req SettlementMarketPayOrderRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail[SettlementMarketPayOrderResponseDTO](enums.ILLEGAL_PARAMETER.Code, enums.ILLEGAL_PARAMETER.Info))
		return
	}
	if req.UserID == "" || req.Source == "" || req.Channel == "" || req.OutTradeNo == "" || req.OutTradeTime.IsZero() {
		ctx.JSON(http.StatusOK, response.Fail[SettlementMarketPayOrderResponseDTO](enums.ILLEGAL_PARAMETER.Code, enums.ILLEGAL_PARAMETER.Info))
		return
	}

	slog.Info("营销交易组队结算开始", "userId", req.UserID, "outTradeNo", req.OutTradeNo)
	result, err := c.settlementService.SettlementMarketPayOrder(ctx.Request.Context(), &entity.TradePaySuccessEntity{
		Source:       req.Source,
		Channel:      req.Channel,
		UserID:       req.UserID,
		OutTradeNo:   req.OutTradeNo,
		OutTradeTime: req.OutTradeTime,
	})
	if err != nil {
		if ae, ok := exception.AsAppException(err); ok {
			ctx.JSON(http.StatusOK, response.Fail[SettlementMarketPayOrderResponseDTO](ae.Code, ae.Info))
			return
		}
		slog.Error("结算失败", "err", err)
		ctx.JSON(http.StatusOK, response.Fail[SettlementMarketPayOrderResponseDTO](enums.UN_ERROR.Code, enums.UN_ERROR.Info))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(SettlementMarketPayOrderResponseDTO{
		UserID:     result.UserID,
		TeamID:     result.TeamID,
		ActivityID: result.ActivityID,
		OutTradeNo: result.OutTradeNo,
	}))
}

func (c *MarketTradeController) RefundMarketPayOrder(ctx *gin.Context) {
	var req RefundMarketPayOrderRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail[RefundMarketPayOrderResponseDTO](enums.ILLEGAL_PARAMETER.Code, enums.ILLEGAL_PARAMETER.Info))
		return
	}
	if req.UserID == "" || req.OutTradeNo == "" || req.Source == "" || req.Channel == "" {
		ctx.JSON(http.StatusOK, response.Fail[RefundMarketPayOrderResponseDTO](enums.ILLEGAL_PARAMETER.Code, enums.ILLEGAL_PARAMETER.Info))
		return
	}

	slog.Info("营销拼团退单开始", "userId", req.UserID, "outTradeNo", req.OutTradeNo)
	result, err := c.refundService.RefundOrder(ctx.Request.Context(), &entity.TradeRefundCommandEntity{
		UserID:     req.UserID,
		OutTradeNo: req.OutTradeNo,
		Source:     req.Source,
		Channel:    req.Channel,
	})
	if err != nil {
		if ae, ok := exception.AsAppException(err); ok {
			ctx.JSON(http.StatusOK, response.Fail[RefundMarketPayOrderResponseDTO](ae.Code, ae.Info))
			return
		}
		slog.Error("退单失败", "err", err)
		ctx.JSON(http.StatusOK, response.Fail[RefundMarketPayOrderResponseDTO](enums.UN_ERROR.Code, enums.UN_ERROR.Info))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(RefundMarketPayOrderResponseDTO{
		UserID:  result.UserID,
		OrderID: result.OrderID,
		TeamID:  result.TeamID,
		Code:    result.Behavior.Code(),
		Info:    result.Behavior.Info(),
	}))
}

func (c *MarketTradeController) failLock(ctx *gin.Context, err error) {
	if ae, ok := exception.AsAppException(err); ok {
		ctx.JSON(http.StatusOK, response.Fail[LockMarketPayOrderResponseDTO](ae.Code, ae.Info))
		return
	}
	slog.Error("锁单失败", "err", err)
	ctx.JSON(http.StatusOK, response.Fail[LockMarketPayOrderResponseDTO](enums.UN_ERROR.Code, enums.UN_ERROR.Info))
}

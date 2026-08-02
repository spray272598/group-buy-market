package http

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"group-buy-market/internal/domain/activity/model/entity"
	"group-buy-market/internal/domain/activity/service"
	"group-buy-market/internal/types/enums"
	"group-buy-market/internal/types/exception"
	"group-buy-market/internal/types/response"
)

// MarketIndexController 营销首页 Trigger
type MarketIndexController struct {
	indexService service.IIndexGroupBuyMarketService
	rateLimit    *RateLimitStore
}

func NewMarketIndexController(indexService service.IIndexGroupBuyMarketService, rl *RateLimitStore) *MarketIndexController {
	return &MarketIndexController{indexService: indexService, rateLimit: rl}
}

func (c *MarketIndexController) Register(r *gin.Engine) {
	g := r.Group("/api/v1/gbm/index")
	g.POST("/query_group_buy_market_config", c.QueryGroupBuyMarketConfig)
}

func (c *MarketIndexController) QueryGroupBuyMarketConfig(ctx *gin.Context) {
	// 读取 body 以便限流键提取 + 绑定
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail[GoodsMarketResponseDTO](enums.ILLEGAL_PARAMETER.Code, enums.ILLEGAL_PARAMETER.Info))
		return
	}
	ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	var req GoodsMarketRequestDTO
	if err := json.Unmarshal(body, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail[GoodsMarketResponseDTO](enums.ILLEGAL_PARAMETER.Code, enums.ILLEGAL_PARAMETER.Info))
		return
	}

	// 按 userId 限流（对齐 Java @RateLimiterAccessInterceptor）
	if c.rateLimit != nil && !c.rateLimit.allow(req.UserID) {
		slog.Error("查询拼团营销配置限流", "userId", req.UserID)
		ctx.JSON(http.StatusOK, response.Fail[GoodsMarketResponseDTO](enums.RATE_LIMITER.Code, enums.RATE_LIMITER.Info))
		return
	}

	if req.UserID == "" || req.Source == "" || req.Channel == "" || req.GoodsID == "" {
		ctx.JSON(http.StatusOK, response.Fail[GoodsMarketResponseDTO](enums.ILLEGAL_PARAMETER.Code, enums.ILLEGAL_PARAMETER.Info))
		return
	}

	slog.Info("查询拼团营销配置开始", "userId", req.UserID, "goodsId", req.GoodsID)

	trial, err := c.indexService.IndexMarketTrial(ctx.Request.Context(), &entity.MarketProductEntity{
		UserID:  req.UserID,
		Source:  req.Source,
		Channel: req.Channel,
		GoodsID: req.GoodsID,
	})
	if err != nil {
		if ae, ok := exception.AsAppException(err); ok {
			ctx.JSON(http.StatusOK, response.Fail[GoodsMarketResponseDTO](ae.Code, ae.Info))
			return
		}
		slog.Error("查询拼团营销配置失败", "err", err)
		ctx.JSON(http.StatusOK, response.Fail[GoodsMarketResponseDTO](enums.UN_ERROR.Code, enums.UN_ERROR.Info))
		return
	}

	activityID := trial.GroupBuyActivityDiscountVO.ActivityID
	teams, err := c.indexService.QueryInProgressUserGroupBuyOrderDetailList(ctx.Request.Context(), activityID, req.UserID, 1, 2)
	if err != nil {
		slog.Error("查询拼团组队失败", "err", err)
		ctx.JSON(http.StatusOK, response.Fail[GoodsMarketResponseDTO](enums.UN_ERROR.Code, enums.UN_ERROR.Info))
		return
	}
	stat, err := c.indexService.QueryTeamStatisticByActivityID(ctx.Request.Context(), activityID)
	if err != nil {
		slog.Error("查询拼团统计失败", "err", err)
		ctx.JSON(http.StatusOK, response.Fail[GoodsMarketResponseDTO](enums.UN_ERROR.Code, enums.UN_ERROR.Info))
		return
	}

	teamDTOs := make([]TeamDTO, 0, len(teams))
	now := time.Now()
	for _, t := range teams {
		teamDTOs = append(teamDTOs, TeamDTO{
			UserID:             t.UserID,
			TeamID:             t.TeamID,
			ActivityID:         t.ActivityID,
			TargetCount:        t.TargetCount,
			CompleteCount:      t.CompleteCount,
			LockCount:          t.LockCount,
			ValidStartTime:     t.ValidStartTime,
			ValidEndTime:       t.ValidEndTime,
			ValidTimeCountdown: CountdownStr(now, t.ValidEndTime),
			OutTradeNo:         t.OutTradeNo,
		})
	}

	data := GoodsMarketResponseDTO{
		ActivityID: activityID,
		Goods: &GoodsDTO{
			GoodsID:        trial.GoodsID,
			OriginalPrice:  trial.OriginalPrice,
			DeductionPrice: trial.DeductionPrice,
			PayPrice:       trial.PayPrice,
		},
		TeamList: teamDTOs,
		TeamStatistic: &TeamStatDTO{
			AllTeamCount:         stat.AllTeamCount,
			AllTeamCompleteCount: stat.AllTeamCompleteCount,
			AllTeamUserCount:     stat.AllTeamUserCount,
		},
	}
	slog.Info("查询拼团营销配置完成", "userId", req.UserID, "activityId", activityID)
	ctx.JSON(http.StatusOK, response.Success(data))
}

package assembler

import (
	"time"

	"group-buy-market/internal/api/dto"
	activityentity "group-buy-market/internal/domain/activity/model/entity"
	tradeentity "group-buy-market/internal/domain/trade/model/entity"
	"group-buy-market/internal/domain/trade/model/valobj"
)

// ToMarketProduct 入站防腐：营销查询 DTO → 领域试算请求
func ToMarketProduct(req *dto.GoodsMarketRequestDTO) *activityentity.MarketProductEntity {
	if req == nil {
		return nil
	}
	return &activityentity.MarketProductEntity{
		UserID:  req.UserID,
		Source:  req.Source,
		Channel: req.Channel,
		GoodsID: req.GoodsID,
	}
}

// ToMarketProductWithActivity 锁单场景：带 activityId
func ToMarketProductWithActivity(req *dto.LockMarketPayOrderRequestDTO) *activityentity.MarketProductEntity {
	if req == nil {
		return nil
	}
	id := req.ActivityID
	return &activityentity.MarketProductEntity{
		UserID:     req.UserID,
		Source:     req.Source,
		Channel:    req.Channel,
		GoodsID:    req.GoodsID,
		ActivityID: &id,
	}
}

// ToGoodsMarketResponse 出站防腐：试算/组队/统计 → 首页响应 DTO
func ToGoodsMarketResponse(
	trial *activityentity.TrialBalanceEntity,
	teams []*activityentity.UserGroupBuyOrderDetailEntity,
	allTeam, allComplete, allUser int64,
) dto.GoodsMarketResponseDTO {
	now := time.Now()
	teamDTOs := make([]dto.TeamDTO, 0, len(teams))
	for _, t := range teams {
		if t == nil {
			continue
		}
		teamDTOs = append(teamDTOs, dto.TeamDTO{
			UserID:             t.UserID,
			TeamID:             t.TeamID,
			ActivityID:         t.ActivityID,
			TargetCount:        t.TargetCount,
			CompleteCount:      t.CompleteCount,
			LockCount:          t.LockCount,
			ValidStartTime:     t.ValidStartTime,
			ValidEndTime:       t.ValidEndTime,
			ValidTimeCountdown: dto.CountdownStr(now, t.ValidEndTime),
			OutTradeNo:         t.OutTradeNo,
		})
	}
	var activityID int64
	var goods *dto.GoodsDTO
	if trial != nil {
		if trial.GroupBuyActivityDiscountVO != nil {
			activityID = trial.GroupBuyActivityDiscountVO.ActivityID
		}
		goods = &dto.GoodsDTO{
			GoodsID:        trial.GoodsID,
			OriginalPrice:  trial.OriginalPrice,
			DeductionPrice: trial.DeductionPrice,
			PayPrice:       trial.PayPrice,
		}
	}
	return dto.GoodsMarketResponseDTO{
		ActivityID: activityID,
		Goods:      goods,
		TeamList:   teamDTOs,
		TeamStatistic: &dto.TeamStatDTO{
			AllTeamCount:         allTeam,
			AllTeamCompleteCount: allComplete,
			AllTeamUserCount:     allUser,
		},
	}
}

// NormalizeLockNotify 入站：补齐 notify 默认值
func NormalizeLockNotify(req *dto.LockMarketPayOrderRequestDTO) {
	if req == nil {
		return
	}
	if req.NotifyConfig == nil && req.NotifyURL != "" {
		req.NotifyConfig = &dto.NotifyConfigDTO{NotifyType: "HTTP", NotifyUrl: req.NotifyURL}
	}
	if req.NotifyConfig == nil {
		req.NotifyConfig = &dto.NotifyConfigDTO{NotifyType: "MQ"}
	}
}

// ToLockResponse 领域锁单结果 → DTO
func ToLockResponse(order *tradeentity.MarketPayOrderEntity) dto.LockMarketPayOrderResponseDTO {
	if order == nil {
		return dto.LockMarketPayOrderResponseDTO{}
	}
	return dto.LockMarketPayOrderResponseDTO{
		OrderID:          order.OrderID,
		OriginalPrice:    order.OriginalPrice,
		DeductionPrice:   order.DeductionPrice,
		PayPrice:         order.PayPrice,
		TradeOrderStatus: int(order.TradeOrderStatus),
		TeamID:           order.TeamID,
	}
}

// ToNotifyConfigVO DTO → 领域通知配置
func ToNotifyConfigVO(cfg *dto.NotifyConfigDTO) *valobj.NotifyConfigVO {
	if cfg == nil {
		return &valobj.NotifyConfigVO{NotifyType: valobj.NotifyMQ}
	}
	return &valobj.NotifyConfigVO{
		NotifyType: valobj.NotifyType(cfg.NotifyType),
		NotifyMQ:   cfg.NotifyMQ,
		NotifyUrl:  cfg.NotifyUrl,
	}
}

// ToTradePaySuccess 结算 DTO → 领域
func ToTradePaySuccess(req *dto.SettlementMarketPayOrderRequestDTO) *tradeentity.TradePaySuccessEntity {
	return &tradeentity.TradePaySuccessEntity{
		Source:       req.Source,
		Channel:      req.Channel,
		UserID:       req.UserID,
		OutTradeNo:   req.OutTradeNo,
		OutTradeTime: req.OutTradeTime,
	}
}

// ToRefundCommand 退单 DTO → 领域
func ToRefundCommand(req *dto.RefundMarketPayOrderRequestDTO) *tradeentity.TradeRefundCommandEntity {
	return &tradeentity.TradeRefundCommandEntity{
		UserID:     req.UserID,
		OutTradeNo: req.OutTradeNo,
		Source:     req.Source,
		Channel:    req.Channel,
	}
}

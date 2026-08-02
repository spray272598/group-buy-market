package application

import (
	"context"
	"log/slog"

	"group-buy-market/internal/api"
	"group-buy-market/internal/api/dto"
	"group-buy-market/internal/api/response"
	"group-buy-market/internal/application/assembler"
	activityservice "group-buy-market/internal/domain/activity/service"
	"group-buy-market/internal/domain/trade/model/entity"
	"group-buy-market/internal/domain/trade/model/valobj"
	"group-buy-market/internal/domain/trade/service/lock"
	"group-buy-market/internal/domain/trade/service/refund"
	"group-buy-market/internal/domain/trade/service/settlement"
	"group-buy-market/internal/types/enums"
	"group-buy-market/internal/types/exception"
)

var _ api.IMarketTradeService = (*MarketTradeAppService)(nil)

// MarketTradeAppService 交易用例：试算→锁单、结算、退单编排
type MarketTradeAppService struct {
	index      activityservice.IIndexGroupBuyMarketService
	lock       lock.ITradeLockOrderService
	settlement settlement.ITradeSettlementOrderService
	refund     refund.ITradeRefundOrderService
}

func NewMarketTradeAppService(
	index activityservice.IIndexGroupBuyMarketService,
	lockSvc lock.ITradeLockOrderService,
	settlementSvc settlement.ITradeSettlementOrderService,
	refundSvc refund.ITradeRefundOrderService,
) *MarketTradeAppService {
	return &MarketTradeAppService{index: index, lock: lockSvc, settlement: settlementSvc, refund: refundSvc}
}

func (s *MarketTradeAppService) LockMarketPayOrder(ctx context.Context, req *dto.LockMarketPayOrderRequestDTO) response.Response[dto.LockMarketPayOrderResponseDTO] {
	if req == nil {
		return response.Fail[dto.LockMarketPayOrderResponseDTO](enums.ILLEGAL_PARAMETER.Code, enums.ILLEGAL_PARAMETER.Info)
	}
	assembler.NormalizeLockNotify(req)
	if req.UserID == "" || req.Source == "" || req.Channel == "" || req.GoodsID == "" || req.ActivityID == 0 ||
		(req.NotifyConfig.NotifyType == "HTTP" && req.NotifyConfig.NotifyUrl == "") {
		return response.Fail[dto.LockMarketPayOrderResponseDTO](enums.ILLEGAL_PARAMETER.Code, enums.ILLEGAL_PARAMETER.Info)
	}

	slog.Info("营销交易锁单", "userId", req.UserID, "outTradeNo", req.OutTradeNo)

	exist, err := s.lock.QueryNoPayMarketPayOrderByOutTradeNo(ctx, req.UserID, req.OutTradeNo)
	if err != nil {
		return failLock(err)
	}
	if exist != nil && exist.TradeOrderStatus == valobj.TradeOrderCreate {
		return response.Success(assembler.ToLockResponse(exist))
	}

	if req.TeamID != "" {
		progress, err := s.lock.QueryGroupBuyProgress(ctx, req.TeamID)
		if err != nil {
			return failLock(err)
		}
		if progress != nil && progress.TargetCount == progress.LockCount {
			return response.Fail[dto.LockMarketPayOrderResponseDTO](enums.E0006.Code, enums.E0006.Info)
		}
	}

	// 用例编排：试算（activity）+ 锁单（trade）
	trial, err := s.index.IndexMarketTrial(ctx, assembler.ToMarketProductWithActivity(req))
	if err != nil {
		return failLock(err)
	}
	if !trial.IsVisible || !trial.IsEnable {
		return response.Fail[dto.LockMarketPayOrderResponseDTO](enums.E0007.Code, enums.E0007.Info)
	}

	g := trial.GroupBuyActivityDiscountVO
	order, err := s.lock.LockMarketPayOrder(ctx,
		&entity.UserEntity{UserID: req.UserID},
		&entity.PayActivityEntity{
			TeamID: req.TeamID, ActivityID: req.ActivityID, ActivityName: g.ActivityName,
			StartTime: g.StartTime, EndTime: g.EndTime, ValidTime: g.ValidTime, TargetCount: g.Target,
		},
		&entity.PayDiscountEntity{
			Source: req.Source, Channel: req.Channel, GoodsID: req.GoodsID, GoodsName: trial.GoodsName,
			OriginalPrice: trial.OriginalPrice, DeductionPrice: trial.DeductionPrice, PayPrice: trial.PayPrice,
			OutTradeNo: req.OutTradeNo, NotifyConfig: assembler.ToNotifyConfigVO(req.NotifyConfig),
		},
	)
	if err != nil {
		return failLock(err)
	}
	return response.Success(assembler.ToLockResponse(order))
}

func (s *MarketTradeAppService) SettlementMarketPayOrder(ctx context.Context, req *dto.SettlementMarketPayOrderRequestDTO) response.Response[dto.SettlementMarketPayOrderResponseDTO] {
	if req == nil || req.UserID == "" || req.Source == "" || req.Channel == "" || req.OutTradeNo == "" || req.OutTradeTime.IsZero() {
		return response.Fail[dto.SettlementMarketPayOrderResponseDTO](enums.ILLEGAL_PARAMETER.Code, enums.ILLEGAL_PARAMETER.Info)
	}
	result, err := s.settlement.SettlementMarketPayOrder(ctx, assembler.ToTradePaySuccess(req))
	if err != nil {
		if ae, ok := exception.AsAppException(err); ok {
			return response.Fail[dto.SettlementMarketPayOrderResponseDTO](ae.Code, ae.Info)
		}
		return response.Fail[dto.SettlementMarketPayOrderResponseDTO](enums.UN_ERROR.Code, enums.UN_ERROR.Info)
	}
	return response.Success(dto.SettlementMarketPayOrderResponseDTO{
		UserID: result.UserID, TeamID: result.TeamID, ActivityID: result.ActivityID, OutTradeNo: result.OutTradeNo,
	})
}

func (s *MarketTradeAppService) RefundMarketPayOrder(ctx context.Context, req *dto.RefundMarketPayOrderRequestDTO) response.Response[dto.RefundMarketPayOrderResponseDTO] {
	if req == nil || req.UserID == "" || req.OutTradeNo == "" || req.Source == "" || req.Channel == "" {
		return response.Fail[dto.RefundMarketPayOrderResponseDTO](enums.ILLEGAL_PARAMETER.Code, enums.ILLEGAL_PARAMETER.Info)
	}
	result, err := s.refund.RefundOrder(ctx, assembler.ToRefundCommand(req))
	if err != nil {
		if ae, ok := exception.AsAppException(err); ok {
			return response.Fail[dto.RefundMarketPayOrderResponseDTO](ae.Code, ae.Info)
		}
		return response.Fail[dto.RefundMarketPayOrderResponseDTO](enums.UN_ERROR.Code, enums.UN_ERROR.Info)
	}
	return response.Success(dto.RefundMarketPayOrderResponseDTO{
		UserID: result.UserID, OrderID: result.OrderID, TeamID: result.TeamID,
		Code: result.Behavior.Code(), Info: result.Behavior.Info(),
	})
}

func failLock(err error) response.Response[dto.LockMarketPayOrderResponseDTO] {
	if ae, ok := exception.AsAppException(err); ok {
		return response.Fail[dto.LockMarketPayOrderResponseDTO](ae.Code, ae.Info)
	}
	slog.Error("锁单失败", "err", err)
	return response.Fail[dto.LockMarketPayOrderResponseDTO](enums.UN_ERROR.Code, enums.UN_ERROR.Info)
}

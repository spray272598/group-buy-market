package application

import (
	"context"

	"group-buy-market/internal/api"
	"group-buy-market/internal/api/dto"
	"group-buy-market/internal/api/response"
	"group-buy-market/internal/application/assembler"
	activityservice "group-buy-market/internal/domain/activity/service"
	"group-buy-market/internal/types/enums"
	"group-buy-market/internal/types/exception"
)

var _ api.IMarketIndexService = (*MarketIndexAppService)(nil)

// MarketIndexAppService 首页营销用例（应用层）
type MarketIndexAppService struct {
	index activityservice.IIndexGroupBuyMarketService
}

func NewMarketIndexAppService(index activityservice.IIndexGroupBuyMarketService) *MarketIndexAppService {
	return &MarketIndexAppService{index: index}
}

func (s *MarketIndexAppService) QueryGroupBuyMarketConfig(ctx context.Context, req *dto.GoodsMarketRequestDTO) response.Response[dto.GoodsMarketResponseDTO] {
	if req == nil || req.UserID == "" || req.Source == "" || req.Channel == "" || req.GoodsID == "" {
		return response.Fail[dto.GoodsMarketResponseDTO](enums.ILLEGAL_PARAMETER.Code, enums.ILLEGAL_PARAMETER.Info)
	}

	// 入站防腐
	product := assembler.ToMarketProduct(req)
	trial, err := s.index.IndexMarketTrial(ctx, product)
	if err != nil {
		if ae, ok := exception.AsAppException(err); ok {
			return response.Fail[dto.GoodsMarketResponseDTO](ae.Code, ae.Info)
		}
		return response.Fail[dto.GoodsMarketResponseDTO](enums.UN_ERROR.Code, enums.UN_ERROR.Info)
	}

	activityID := trial.GroupBuyActivityDiscountVO.ActivityID
	teams, err := s.index.QueryInProgressUserGroupBuyOrderDetailList(ctx, activityID, req.UserID, 1, 2)
	if err != nil {
		return response.Fail[dto.GoodsMarketResponseDTO](enums.UN_ERROR.Code, enums.UN_ERROR.Info)
	}
	stat, err := s.index.QueryTeamStatisticByActivityID(ctx, activityID)
	if err != nil {
		return response.Fail[dto.GoodsMarketResponseDTO](enums.UN_ERROR.Code, enums.UN_ERROR.Info)
	}

	var allTeam, allComplete, allUser int64
	if stat != nil {
		allTeam, allComplete, allUser = stat.AllTeamCount, stat.AllTeamCompleteCount, stat.AllTeamUserCount
	}
	// 出站防腐
	return response.Success(assembler.ToGoodsMarketResponse(trial, teams, allTeam, allComplete, allUser))
}

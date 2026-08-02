package service

import (
	"context"

	"group-buy-market/internal/domain/activity/adapter/repository"
	"group-buy-market/internal/domain/activity/model/entity"
	"group-buy-market/internal/domain/activity/model/valobj"
	"group-buy-market/internal/domain/activity/service/trial/node"
)

// IIndexGroupBuyMarketService 首页营销服务
type IIndexGroupBuyMarketService interface {
	IndexMarketTrial(ctx context.Context, product *entity.MarketProductEntity) (*entity.TrialBalanceEntity, error)
	QueryInProgressUserGroupBuyOrderDetailList(ctx context.Context, activityID int64, userID string, ownerCount, randomCount int) ([]*entity.UserGroupBuyOrderDetailEntity, error)
	QueryTeamStatisticByActivityID(ctx context.Context, activityID int64) (*valobj.TeamStatisticVO, error)
}

type IndexGroupBuyMarketService struct {
	repo  repository.IActivityRepository
	chain *node.Chain
}

func NewIndexGroupBuyMarketService(repo repository.IActivityRepository, chain *node.Chain) *IndexGroupBuyMarketService {
	return &IndexGroupBuyMarketService{repo: repo, chain: chain}
}

func (s *IndexGroupBuyMarketService) IndexMarketTrial(ctx context.Context, product *entity.MarketProductEntity) (*entity.TrialBalanceEntity, error) {
	return s.chain.Apply(ctx, product)
}

func (s *IndexGroupBuyMarketService) QueryInProgressUserGroupBuyOrderDetailList(ctx context.Context, activityID int64, userID string, ownerCount, randomCount int) ([]*entity.UserGroupBuyOrderDetailEntity, error) {
	var result []*entity.UserGroupBuyOrderDetailEntity
	if ownerCount != 0 {
		list, err := s.repo.QueryInProgressUserGroupBuyOrderDetailListByOwner(ctx, activityID, userID, ownerCount)
		if err != nil {
			return nil, err
		}
		result = append(result, list...)
	}
	if randomCount != 0 {
		list, err := s.repo.QueryInProgressUserGroupBuyOrderDetailListByRandom(ctx, activityID, userID, randomCount)
		if err != nil {
			return nil, err
		}
		result = append(result, list...)
	}
	return result, nil
}

func (s *IndexGroupBuyMarketService) QueryTeamStatisticByActivityID(ctx context.Context, activityID int64) (*valobj.TeamStatisticVO, error) {
	return s.repo.QueryTeamStatisticByActivityID(ctx, activityID)
}

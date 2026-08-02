package factory

import (
	"group-buy-market/internal/domain/trade/model/entity"
	"group-buy-market/internal/types/common"
)

// DynamicContext 锁单规则链上下文
type DynamicContext struct {
	GroupBuyActivity   *entity.GroupBuyActivityEntity
	UserTakeOrderCount int
}

func (d *DynamicContext) GenerateTeamStockKey(teamID string) string {
	if teamID == "" || d.GroupBuyActivity == nil {
		return ""
	}
	return GenerateTeamStockKey(d.GroupBuyActivity.ActivityID, teamID)
}

func (d *DynamicContext) GenerateRecoveryTeamStockKey(teamID string) string {
	if teamID == "" || d.GroupBuyActivity == nil {
		return ""
	}
	return GenerateRecoveryTeamStockKey(d.GroupBuyActivity.ActivityID, teamID)
}

func GenerateTeamStockKey(activityID int64, teamID string) string {
	return common.TeamStockKeyPrefix + itoa(activityID) + common.Underline + teamID
}

func GenerateRecoveryTeamStockKey(activityID int64, teamID string) string {
	return common.TeamStockKeyPrefix + itoa(activityID) + common.Underline + teamID + "_recovery"
}

func itoa(v int64) string {
	// 避免额外依赖
	const digits = "0123456789"
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = digits[v%10]
		v /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

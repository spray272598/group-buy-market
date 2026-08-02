package valobj

import (
	"strings"
	"time"
)

// GroupBuyActivityDiscountVO 拼团活动营销配置值对象
type GroupBuyActivityDiscountVO struct {
	ActivityID       int64
	ActivityName     string
	Source           string
	Channel          string
	GoodsID          string
	GroupBuyDiscount *GroupBuyDiscount
	GroupType        int
	TakeLimitCount   int
	Target           int
	ValidTime        int // 分钟
	Status           int
	StartTime        time.Time
	EndTime          time.Time
	TagID            string
	TagScope         string // 人群标签规则范围
}

// GroupBuyDiscount 折扣配置
type GroupBuyDiscount struct {
	DiscountName string
	DiscountDesc string
	DiscountType DiscountType
	MarketPlan   string // ZJ/MJ/N/ZK
	MarketExpr   string
	TagID        string
}

// IsVisible 可见限制：tagScope 含 "1" 时默认不可见，需人群命中
func (v *GroupBuyActivityDiscountVO) IsVisible() bool {
	if v.TagScope == "" {
		return true
	}
	parts := strings.Split(v.TagScope, ",")
	if len(parts) > 0 && parts[0] == "1" {
		return false
	}
	return true
}

// IsEnable 参与限制：tagScope 含 "2" 时默认不可参与，需人群命中
func (v *GroupBuyActivityDiscountVO) IsEnable() bool {
	if v.TagScope == "" {
		return true
	}
	parts := strings.Split(v.TagScope, ",")
	if len(parts) == 2 && parts[1] == "2" {
		return false
	}
	if len(parts) == 1 && parts[0] == "2" {
		return false
	}
	return true
}

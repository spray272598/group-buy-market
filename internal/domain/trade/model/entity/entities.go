package entity

import (
	"time"

	"github.com/shopspring/decimal"

	"group-buy-market/internal/domain/trade/model/valobj"
	"group-buy-market/internal/types/enums"
)

// UserEntity 用户
type UserEntity struct {
	UserID string
}

// PayActivityEntity 支付活动信息
type PayActivityEntity struct {
	TeamID       string
	ActivityID   int64
	ActivityName string
	StartTime    time.Time
	EndTime      time.Time
	ValidTime    int // 分钟
	TargetCount  int
}

// PayDiscountEntity 支付优惠信息
type PayDiscountEntity struct {
	Source         string
	Channel        string
	GoodsID        string
	GoodsName      string
	OriginalPrice  decimal.Decimal
	DeductionPrice decimal.Decimal
	PayPrice       decimal.Decimal
	OutTradeNo     string
	NotifyConfig   *valobj.NotifyConfigVO
}

// MarketPayOrderEntity 营销支付订单
type MarketPayOrderEntity struct {
	TeamID            string
	OrderID           string
	OriginalPrice     decimal.Decimal
	DeductionPrice    decimal.Decimal
	PayPrice          decimal.Decimal
	TradeOrderStatus  valobj.TradeOrderStatus
}

// GroupBuyActivityEntity 拼团活动（交易域）
type GroupBuyActivityEntity struct {
	ActivityID     int64
	ActivityName   string
	DiscountID     string
	GroupType      int
	TakeLimitCount int
	Target         int
	ValidTime      int
	Status         enums.ActivityStatus
	StartTime      time.Time
	EndTime        time.Time
	TagID          string
	TagScope       string
}

// GroupBuyTeamEntity 拼团组队
type GroupBuyTeamEntity struct {
	TeamID         string
	ActivityID     int64
	TargetCount    int
	CompleteCount  int
	LockCount      int
	Status         enums.GroupBuyOrderStatus
	ValidStartTime time.Time
	ValidEndTime   time.Time
	NotifyConfig   *valobj.NotifyConfigVO
}

// NotifyTaskEntity 回调任务
type NotifyTaskEntity struct {
	TeamID        string
	NotifyType    string
	NotifyMQ      string
	NotifyUrl     string
	NotifyCount   int
	Status        valobj.NotifyTaskStatus
	ParameterJSON string
	UUID          string
	ActivityID    int64
}

// LockKey 多实例回调抢锁 key
func (n *NotifyTaskEntity) LockKey() string {
	return "notify_job_lock_key_" + n.UUID
}

// TradeLockRuleCommandEntity 锁单规则命令
type TradeLockRuleCommandEntity struct {
	UserID     string
	ActivityID int64
	TeamID     string
}

// TradeLockRuleFilterBackEntity 锁单规则结果
type TradeLockRuleFilterBackEntity struct {
	UserTakeOrderCount  int
	RecoveryTeamStockKey string
}

// TradePaySuccessEntity 支付成功
type TradePaySuccessEntity struct {
	Source       string
	Channel      string
	UserID       string
	OutTradeNo   string
	OutTradeTime time.Time
}

// TradePaySettlementEntity 结算结果
type TradePaySettlementEntity struct {
	Source     string
	Channel    string
	UserID     string
	TeamID     string
	ActivityID int64
	OutTradeNo string
}

// TradeSettlementRuleCommandEntity 结算规则命令
type TradeSettlementRuleCommandEntity struct {
	Source       string
	Channel      string
	UserID       string
	OutTradeNo   string
	OutTradeTime time.Time
}

// TradeSettlementRuleFilterBackEntity 结算规则结果
type TradeSettlementRuleFilterBackEntity struct {
	TeamID         string
	ActivityID     int64
	TargetCount    int
	CompleteCount  int
	LockCount      int
	Status         enums.GroupBuyOrderStatus
	ValidStartTime time.Time
	ValidEndTime   time.Time
	NotifyConfig   *valobj.NotifyConfigVO
}

// TradeRefundCommandEntity 退单命令
type TradeRefundCommandEntity struct {
	UserID     string
	OutTradeNo string
	Source     string
	Channel    string
}

// TradeRefundOrderEntity 退单订单
type TradeRefundOrderEntity struct {
	UserID     string
	OrderID    string
	TeamID     string
	ActivityID int64
	OutTradeNo string
}

// TradeRefundBehavior 退单行为
type TradeRefundBehavior int

const (
	RefundBehaviorSuccess TradeRefundBehavior = 1
	RefundBehaviorRepeat  TradeRefundBehavior = 2
)

func (b TradeRefundBehavior) Code() string {
	switch b {
	case RefundBehaviorSuccess:
		return "success"
	case RefundBehaviorRepeat:
		return "repeat"
	default:
		return "unknown"
	}
}

func (b TradeRefundBehavior) Info() string {
	switch b {
	case RefundBehaviorSuccess:
		return "退单成功"
	case RefundBehaviorRepeat:
		return "重复退单"
	default:
		return "未知"
	}
}

// TradeRefundBehaviorEntity 退单行为结果
type TradeRefundBehaviorEntity struct {
	UserID   string
	OrderID  string
	TeamID   string
	Behavior TradeRefundBehavior
}

package http

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// GoodsMarketRequestDTO 首页营销配置请求
type GoodsMarketRequestDTO struct {
	UserID  string `json:"userId"`
	Source  string `json:"source"`
	Channel string `json:"channel"`
	GoodsID string `json:"goodsId"`
}

// GoodsMarketResponseDTO 首页营销配置响应
type GoodsMarketResponseDTO struct {
	ActivityID    int64         `json:"activityId"`
	Goods         *GoodsDTO     `json:"goods"`
	TeamList      []TeamDTO     `json:"teamList"`
	TeamStatistic *TeamStatDTO  `json:"teamStatistic"`
}

type GoodsDTO struct {
	GoodsID        string          `json:"goodsId"`
	OriginalPrice  decimal.Decimal `json:"originalPrice"`
	DeductionPrice decimal.Decimal `json:"deductionPrice"`
	PayPrice       decimal.Decimal `json:"payPrice"`
}

type TeamDTO struct {
	UserID             string    `json:"userId"`
	TeamID             string    `json:"teamId"`
	ActivityID         int64     `json:"activityId"`
	TargetCount        int       `json:"targetCount"`
	CompleteCount      int       `json:"completeCount"`
	LockCount          int       `json:"lockCount"`
	ValidStartTime     time.Time `json:"validStartTime"`
	ValidEndTime       time.Time `json:"validEndTime"`
	ValidTimeCountdown string    `json:"validTimeCountdown"`
	OutTradeNo         string    `json:"outTradeNo"`
}

type TeamStatDTO struct {
	AllTeamCount         int64 `json:"allTeamCount"`
	AllTeamCompleteCount int64 `json:"allTeamCompleteCount"`
	AllTeamUserCount     int64 `json:"allTeamUserCount"`
}

func CountdownStr(from, to time.Time) string {
	d := to.Sub(from)
	if d < 0 {
		return "00:00:00"
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

// LockMarketPayOrderRequestDTO 锁单请求
type LockMarketPayOrderRequestDTO struct {
	UserID       string          `json:"userId"`
	TeamID       string          `json:"teamId"`
	ActivityID   int64           `json:"activityId"`
	GoodsID      string          `json:"goodsId"`
	Source       string          `json:"source"`
	Channel      string          `json:"channel"`
	OutTradeNo   string          `json:"outTradeNo"`
	NotifyConfig *NotifyConfigDTO `json:"notifyConfigVO"`
	// 兼容字段
	NotifyURL string `json:"notifyUrl"`
}

type NotifyConfigDTO struct {
	NotifyType string `json:"notifyType"`
	NotifyMQ   string `json:"notifyMQ"`
	NotifyUrl  string `json:"notifyUrl"`
}

// LockMarketPayOrderResponseDTO 锁单响应
type LockMarketPayOrderResponseDTO struct {
	OrderID          string          `json:"orderId"`
	OriginalPrice    decimal.Decimal `json:"originalPrice"`
	DeductionPrice   decimal.Decimal `json:"deductionPrice"`
	PayPrice         decimal.Decimal `json:"payPrice"`
	TradeOrderStatus int             `json:"tradeOrderStatus"`
	TeamID           string          `json:"teamId"`
}

// SettlementMarketPayOrderRequestDTO 结算请求
type SettlementMarketPayOrderRequestDTO struct {
	UserID       string    `json:"userId"`
	Source       string    `json:"source"`
	Channel      string    `json:"channel"`
	OutTradeNo   string    `json:"outTradeNo"`
	OutTradeTime time.Time `json:"outTradeTime"`
}

// SettlementMarketPayOrderResponseDTO 结算响应
type SettlementMarketPayOrderResponseDTO struct {
	UserID     string `json:"userId"`
	TeamID     string `json:"teamId"`
	ActivityID int64  `json:"activityId"`
	OutTradeNo string `json:"outTradeNo"`
}

// RefundMarketPayOrderRequestDTO 退单请求
type RefundMarketPayOrderRequestDTO struct {
	UserID     string `json:"userId"`
	OutTradeNo string `json:"outTradeNo"`
	Source     string `json:"source"`
	Channel    string `json:"channel"`
}

// RefundMarketPayOrderResponseDTO 退单响应
type RefundMarketPayOrderResponseDTO struct {
	UserID  string `json:"userId"`
	OrderID string `json:"orderId"`
	TeamID  string `json:"teamId"`
	Code    string `json:"code"`
	Info    string `json:"info"`
}

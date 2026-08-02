package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

// LockMarketPayOrderRequestDTO 锁单请求
type LockMarketPayOrderRequestDTO struct {
	UserID       string           `json:"userId"`
	TeamID       string           `json:"teamId"`
	ActivityID   int64            `json:"activityId"`
	GoodsID      string           `json:"goodsId"`
	Source       string           `json:"source"`
	Channel      string           `json:"channel"`
	OutTradeNo   string           `json:"outTradeNo"`
	NotifyConfig *NotifyConfigDTO `json:"notifyConfigVO"`
	// NotifyURL 兼容旧字段：仅传 URL 时视为 HTTP 回调
	NotifyURL string `json:"notifyUrl"`
}

// NotifyConfigDTO 回调配置
type NotifyConfigDTO struct {
	NotifyType string `json:"notifyType"` // HTTP / MQ
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

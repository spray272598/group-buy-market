package api

import (
	"context"

	"group-buy-market/internal/api/dto"
	"group-buy-market/internal/api/response"
)

// IMarketTradeService 营销交易对外服务契约（对齐 Java IMarketTradeService）
type IMarketTradeService interface {
	LockMarketPayOrder(ctx context.Context, req *dto.LockMarketPayOrderRequestDTO) response.Response[dto.LockMarketPayOrderResponseDTO]
	SettlementMarketPayOrder(ctx context.Context, req *dto.SettlementMarketPayOrderRequestDTO) response.Response[dto.SettlementMarketPayOrderResponseDTO]
	RefundMarketPayOrder(ctx context.Context, req *dto.RefundMarketPayOrderRequestDTO) response.Response[dto.RefundMarketPayOrderResponseDTO]
}

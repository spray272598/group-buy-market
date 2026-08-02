package api

import (
	"context"

	"group-buy-market/internal/api/dto"
	"group-buy-market/internal/api/response"
)

// IMarketIndexService 营销首页对外服务契约（对齐 Java IMarketIndexService）
// 由 trigger/http 实现；领域服务不直接暴露给调用方。
// Go 惯例补充 context.Context 作为首参（Java 无此参数）。
type IMarketIndexService interface {
	QueryGroupBuyMarketConfig(ctx context.Context, req *dto.GoodsMarketRequestDTO) response.Response[dto.GoodsMarketResponseDTO]
}

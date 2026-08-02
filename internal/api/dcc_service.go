package api

import (
	"context"

	"group-buy-market/internal/api/dto"
	"group-buy-market/internal/api/response"
)

// IDCCService 动态配置对外契约（对齐 Java IDCCService）
type IDCCService interface {
	UpdateConfig(ctx context.Context, key, value string) response.Response[bool]
	Query(ctx context.Context) response.Response[dto.DCCSnapshot]
}

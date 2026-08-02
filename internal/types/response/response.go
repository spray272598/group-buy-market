// Package response 兼容层：统一响应已迁至 api/response。
// 新代码请使用 group-buy-market/internal/api/response。
package response

import apiresp "group-buy-market/internal/api/response"

// Response 类型别名
type Response[T any] = apiresp.Response[T]

func Success[T any](data T) Response[T] {
	return apiresp.Success(data)
}

func Fail[T any](code, info string) Response[T] {
	return apiresp.Fail[T](code, info)
}

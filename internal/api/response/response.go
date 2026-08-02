package response

// Response 统一 API 响应结构（对齐 Java api.response.Response）
type Response[T any] struct {
	Code string `json:"code"`
	Info string `json:"info"`
	Data T      `json:"data,omitempty"`
}

func Success[T any](data T) Response[T] {
	return Response[T]{Code: "0000", Info: "成功", Data: data}
}

func Fail[T any](code, info string) Response[T] {
	return Response[T]{Code: code, Info: info}
}

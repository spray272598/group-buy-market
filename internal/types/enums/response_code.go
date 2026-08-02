package enums

// ResponseCode 统一业务响应码（对齐 Java ResponseCode）
type ResponseCode struct {
	Code string
	Info string
}

var (
	SUCCESS           = ResponseCode{"0000", "成功"}
	UN_ERROR          = ResponseCode{"0001", "未知失败"}
	ILLEGAL_PARAMETER = ResponseCode{"0002", "非法参数"}
	INDEX_EXCEPTION   = ResponseCode{"0003", "唯一索引冲突"}
	UPDATE_ZERO       = ResponseCode{"0004", "更新记录为0"}
	HTTP_EXCEPTION    = ResponseCode{"0005", "HTTP接口调用异常"}
	RATE_LIMITER      = ResponseCode{"0006", "接口限流"}

	E0001 = ResponseCode{"E0001", "不存在对应的折扣计算服务"}
	E0002 = ResponseCode{"E0002", "无拼团营销配置"}
	E0003 = ResponseCode{"E0003", "拼团活动降级拦截"}
	E0004 = ResponseCode{"E0004", "拼团活动切量拦截"}
	E0005 = ResponseCode{"E0005", "拼团组队失败，记录更新为0"}
	E0006 = ResponseCode{"E0006", "拼团组队完结，锁单量已达成"}
	E0007 = ResponseCode{"E0007", "拼团人群限定，不可参与"}
	E0008 = ResponseCode{"E0008", "拼团组队失败，缓存库存不足"}

	E0101 = ResponseCode{"E0101", "拼团活动未生效"}
	E0102 = ResponseCode{"E0102", "不在拼团活动有效时间内"}
	E0103 = ResponseCode{"E0103", "当前用户参与此拼团次数已达上限"}
	E0104 = ResponseCode{"E0104", "不存在的外部交易单号或用户已退单"}
	E0105 = ResponseCode{"E0105", "SC渠道黑名单拦截"}
	E0106 = ResponseCode{"E0106", "订单交易时间不在拼团有效时间范围内"}
)

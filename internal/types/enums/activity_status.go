package enums

// ActivityStatus 活动状态（0创建、1生效、2过期、3废弃）
type ActivityStatus int

const (
	ActivityCreate    ActivityStatus = 0
	ActivityEffective ActivityStatus = 1
	ActivityExpire    ActivityStatus = 2
	ActivityAbandon   ActivityStatus = 3
)

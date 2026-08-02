package enums

const (
	NotifyTaskHTTPSuccess = "success"
	NotifyTaskHTTPError   = "error"
	NotifyTaskHTTPNull    = "null" // 未抢到锁，不更新任务状态
)

package valobj

import (
	"fmt"
	"slices"
)

// NotifyTaskStatus 回调任务状态
// 0初始 1成功 2重试 3失败
type NotifyTaskStatus int

const (
	NotifyTaskInit    NotifyTaskStatus = 0
	NotifyTaskSuccess NotifyTaskStatus = 1
	NotifyTaskRetry   NotifyTaskStatus = 2
	NotifyTaskError   NotifyTaskStatus = 3
)

// notifyTransitions 状态迁移表：key 为当前状态，value 为允许迁移到的目标状态
var notifyTransitions = map[NotifyTaskStatus][]NotifyTaskStatus{
	NotifyTaskInit:    {NotifyTaskSuccess, NotifyTaskRetry, NotifyTaskError},
	NotifyTaskRetry:   {NotifyTaskSuccess, NotifyTaskError},
	NotifyTaskSuccess: {}, // 终态
	NotifyTaskError:   {}, // 终态
}

// CanTransition 判断当前状态能否迁移到目标状态（非法流转返回 false）
func (s NotifyTaskStatus) CanTransition(to NotifyTaskStatus) bool {
	return slices.Contains(notifyTransitions[s], to)
}

// MoveTo 执行状态迁移；非法迁移返回错误（不修改原状态）
func (s NotifyTaskStatus) MoveTo(to NotifyTaskStatus) (NotifyTaskStatus, error) {
	if !s.CanTransition(to) {
		return s, fmt.Errorf("非法的回调任务状态迁移: %d -> %d", s, to)
	}
	return to, nil
}

// IsTerminal 是否为终态（成功/失败，无需再补偿）
func (s NotifyTaskStatus) IsTerminal() bool {
	return s == NotifyTaskSuccess || s == NotifyTaskError
}

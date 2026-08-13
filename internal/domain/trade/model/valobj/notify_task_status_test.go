package valobj

import "testing"

func TestNotifyTaskStatus_CanTransition(t *testing.T) {
	tests := []struct {
		name string
		from NotifyTaskStatus
		to   NotifyTaskStatus
		want bool
	}{
		{"初始→成功", NotifyTaskInit, NotifyTaskSuccess, true},
		{"初始→重试", NotifyTaskInit, NotifyTaskRetry, true},
		{"初始→失败", NotifyTaskInit, NotifyTaskError, true},
		{"重试→成功", NotifyTaskRetry, NotifyTaskSuccess, true},
		{"重试→失败", NotifyTaskRetry, NotifyTaskError, true},
		{"重试→初始(非法)", NotifyTaskRetry, NotifyTaskInit, false},
		{"成功→重试(非法,终态)", NotifyTaskSuccess, NotifyTaskRetry, false},
		{"成功→失败(非法,终态)", NotifyTaskSuccess, NotifyTaskError, false},
		{"失败→成功(非法,终态)", NotifyTaskError, NotifyTaskSuccess, false},
		{"失败→重试(非法,终态)", NotifyTaskError, NotifyTaskRetry, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.from.CanTransition(tt.to); got != tt.want {
				t.Errorf("CanTransition(%d -> %d) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestNotifyTaskStatus_MoveTo(t *testing.T) {
	// 合法迁移
	if got, err := NotifyTaskInit.MoveTo(NotifyTaskSuccess); err != nil || got != NotifyTaskSuccess {
		t.Errorf("MoveTo 合法迁移失败: got=%d err=%v", got, err)
	}
	// 非法迁移应返回错误且不修改原状态
	if _, err := NotifyTaskSuccess.MoveTo(NotifyTaskRetry); err == nil {
		t.Errorf("终态迁移到重试应返回错误")
	}
}

func TestNotifyTaskStatus_IsTerminal(t *testing.T) {
	if NotifyTaskSuccess.IsTerminal() != true {
		t.Errorf("成功应为终态")
	}
	if NotifyTaskError.IsTerminal() != true {
		t.Errorf("失败应为终态")
	}
	if NotifyTaskInit.IsTerminal() {
		t.Errorf("初始不应为终态")
	}
	if NotifyTaskRetry.IsTerminal() {
		t.Errorf("重试不应为终态")
	}
}

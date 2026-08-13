package task

import (
	"context"
	"log/slog"

	"group-buy-market/internal/domain/trade/adapter/port"
	"group-buy-market/internal/domain/trade/adapter/repository"
	"group-buy-market/internal/domain/trade/model/entity"
	"group-buy-market/internal/domain/trade/model/valobj"
	"group-buy-market/internal/types/enums"
)

// TradeTaskService 回调任务服务（HTTP/MQ），对齐 Java TradeTaskService
type TradeTaskService struct {
	repo repository.ITradeRepository
	port port.ITradePort
}

func NewTradeTaskService(repo repository.ITradeRepository, p port.ITradePort) *TradeTaskService {
	return &TradeTaskService{repo: repo, port: p}
}

func (s *TradeTaskService) ExecNotifyJob(ctx context.Context, task *entity.NotifyTaskEntity) (map[string]int, error) {
	var tasks []*entity.NotifyTaskEntity
	var err error
	if task != nil {
		tasks = []*entity.NotifyTaskEntity{task}
	} else {
		tasks, err = s.repo.QueryUnExecutedNotifyTaskList(ctx)
		if err != nil {
			return nil, err
		}
	}
	return s.execList(ctx, tasks)
}

func (s *TradeTaskService) ExecNotifyJobByTeamID(ctx context.Context, teamID string) (map[string]int, error) {
	tasks, err := s.repo.QueryUnExecutedNotifyTaskListByTeamID(ctx, teamID)
	if err != nil {
		return nil, err
	}
	return s.execList(ctx, tasks)
}

func (s *TradeTaskService) execList(ctx context.Context, tasks []*entity.NotifyTaskEntity) (map[string]int, error) {
	success, errCnt, retry := 0, 0, 0
	for _, t := range tasks {
		// 终态任务直接跳过（已成功/失败），避免重复投递
		if t.Status.IsTerminal() {
			continue
		}
		resp, e := s.port.GroupBuyNotify(ctx, t)
		if e != nil {
			slog.Error("回调通知异常", "teamId", t.TeamID, "err", e)
			resp = enums.NotifyTaskHTTPError
		}
		switch resp {
		case enums.NotifyTaskHTTPSuccess:
			if _, err := t.Status.MoveTo(valobj.NotifyTaskSuccess); err == nil {
				if n, _ := s.repo.UpdateNotifyTaskStatusSuccess(ctx, t); n == 1 {
					success++
				}
			}
		case enums.NotifyTaskHTTPError:
			// 超过 4 次标记失败，否则重试（对齐 Java）
			if t.NotifyCount > 4 {
				if _, err := t.Status.MoveTo(valobj.NotifyTaskError); err == nil {
					if n, _ := s.repo.UpdateNotifyTaskStatusError(ctx, t); n == 1 {
						errCnt++
					}
				}
			} else {
				if _, err := t.Status.MoveTo(valobj.NotifyTaskRetry); err == nil {
					if n, _ := s.repo.UpdateNotifyTaskStatusRetry(ctx, t); n == 1 {
						retry++
					}
				}
			}
		case enums.NotifyTaskHTTPNull:
			// 未抢到锁，不改状态
		default:
			if _, err := t.Status.MoveTo(valobj.NotifyTaskRetry); err == nil {
				if n, _ := s.repo.UpdateNotifyTaskStatusRetry(ctx, t); n == 1 {
					retry++
				}
			}
		}
	}
	return map[string]int{
		"waitCount":    len(tasks),
		"successCount": success,
		"errorCount":   errCnt,
		"retryCount":   retry,
	}, nil
}

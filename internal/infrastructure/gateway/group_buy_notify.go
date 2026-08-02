package gateway

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	"group-buy-market/internal/types/enums"
)

// GroupBuyNotifyService HTTP 回调网关（对齐 Java GroupBuyNotifyService）
type GroupBuyNotifyService struct {
	client *http.Client
}

func NewGroupBuyNotifyService(timeoutSec int) *GroupBuyNotifyService {
	if timeoutSec <= 0 {
		timeoutSec = 5
	}
	return &GroupBuyNotifyService{
		client: &http.Client{Timeout: time.Duration(timeoutSec) * time.Second},
	}
}

func (s *GroupBuyNotifyService) GroupBuyNotify(ctx context.Context, notifyURL, parameterJSON string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, notifyURL, bytes.NewBufferString(parameterJSON))
	if err != nil {
		return enums.NotifyTaskHTTPError, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return enums.NotifyTaskHTTPError, err
	}
	defer resp.Body.Close()
	_, _ = io.ReadAll(resp.Body)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return enums.NotifyTaskHTTPSuccess, nil
	}
	return enums.NotifyTaskHTTPError, nil
}

package dcc

import (
	"context"
	"encoding/json"
	"hash/fnv"
	"log/slog"
	"strings"
	"sync"

	"group-buy-market/internal/infrastructure/metrics"
	"group-buy-market/internal/types/common"
	"group-buy-market/internal/types/safego"
)

// AttributeVO 配置变更消息（对齐 Java AttributeVO，用于 Redis 广播）
type AttributeVO struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Publisher 跨实例发布（由 Redis Pub/Sub 实现）
type Publisher interface {
	Publish(ctx context.Context, channel string, message string) error
}

// Subscriber 跨实例订阅
type Subscriber interface {
	Subscribe(ctx context.Context, channel string, handler func(payload string)) error
}

const DefaultChannel = "group_buy_market_dcc_topic"

// Service 动态配置中心（本地内存 + 可选 Redis 广播）
type Service struct {
	mu              sync.RWMutex
	downgradeSwitch string
	cutRange        string
	scBlacklist     string
	cacheSwitch     string

	publisher Publisher
	channel   string
	// 本机发布时跳过回环应用（订阅侧仍会收到，用 origin 判断）
	instanceID string
}

func New(downgrade, cutRange, scBlacklist, cacheSwitch string) *Service {
	return &Service{
		downgradeSwitch: downgrade,
		cutRange:        cutRange,
		scBlacklist:     scBlacklist,
		cacheSwitch:     cacheSwitch,
		channel:         DefaultChannel,
		instanceID:      "local",
	}
}

// EnableBroadcast 启用 Redis 跨实例通知
func (s *Service) EnableBroadcast(pub Publisher, sub Subscriber, channel, instanceID string) {
	if channel != "" {
		s.channel = channel
	}
	if instanceID != "" {
		s.instanceID = instanceID
	}
	s.publisher = pub
	if sub == nil {
		return
	}
	safego.Go("dcc_subscribe", func() {
		ctx := context.Background()
		err := sub.Subscribe(ctx, s.channel, func(payload string) {
			var msg struct {
				AttributeVO
				Origin string `json:"origin"`
			}
			if err := json.Unmarshal([]byte(payload), &msg); err != nil {
				// 兼容仅 key/value
				var attr AttributeVO
				if err2 := json.Unmarshal([]byte(payload), &attr); err2 != nil {
					slog.Error("DCC 消息解析失败", "payload", payload, "err", err)
					return
				}
				msg.AttributeVO = attr
			}
			// 本实例发出的消息也应用一次无妨（幂等）；日志区分
			slog.Info("DCC 收到跨实例配置变更", "key", msg.Key, "value", msg.Value, "origin", msg.Origin)
			s.applyLocal(msg.Key, msg.Value)
			metrics.ObserveDCC(msg.Key, "broadcast")
		})
		if err != nil {
			slog.Error("DCC 订阅失败", "err", err)
		}
	})
}

func (s *Service) IsDowngradeSwitch() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.downgradeSwitch == "1"
}

func (s *Service) IsCutRange(userID string) bool {
	s.mu.RLock()
	cut := s.cutRange
	s.mu.RUnlock()
	if cut == "" {
		return true
	}
	var rangeN int
	for _, c := range cut {
		if c >= '0' && c <= '9' {
			rangeN = rangeN*10 + int(c-'0')
		}
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(userID))
	lastTwo := int(h.Sum32() % 100)
	return lastTwo <= rangeN
}

func (s *Service) IsSCBlackIntercept(source, channel string) bool {
	s.mu.RLock()
	bl := s.scBlacklist
	s.mu.RUnlock()
	if bl == "" {
		return false
	}
	key := source + channel
	for _, item := range strings.Split(bl, common.Split) {
		if strings.TrimSpace(item) == key {
			return true
		}
	}
	return false
}

func (s *Service) IsCacheOpenSwitch() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cacheSwitch == "0"
}

// Update 本机更新并广播到其他实例
func (s *Service) Update(key, value string) {
	s.applyLocal(key, value)
	metrics.ObserveDCC(key, "local")
	s.broadcast(key, value)
}

// applyLocal 仅改本地内存（订阅端调用）
func (s *Service) applyLocal(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch key {
	case "downgradeSwitch":
		s.downgradeSwitch = value
	case "cutRange":
		s.cutRange = value
	case "scBlacklist":
		s.scBlacklist = value
	case "cacheSwitch":
		s.cacheSwitch = value
	default:
		slog.Warn("DCC 未知配置键", "key", key)
	}
}

func (s *Service) broadcast(key, value string) {
	if s.publisher == nil {
		return
	}
	payload, _ := json.Marshal(map[string]string{
		"key":    key,
		"value":  value,
		"origin": s.instanceID,
	})
	ctx := context.Background()
	if err := s.publisher.Publish(ctx, s.channel, string(payload)); err != nil {
		slog.Error("DCC 广播失败", "key", key, "err", err)
		return
	}
	slog.Info("DCC 已广播配置变更", "key", key, "value", value, "channel", s.channel)
}

func (s *Service) Snapshot() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return map[string]string{
		"downgradeSwitch": s.downgradeSwitch,
		"cutRange":        s.cutRange,
		"scBlacklist":     s.scBlacklist,
		"cacheSwitch":     s.cacheSwitch,
	}
}

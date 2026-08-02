package dcc

import (
	"hash/fnv"
	"strings"
	"sync"

	"group-buy-market/internal/types/common"
)

// Service 动态配置中心（对齐 Java DCCService）
// 领域仓储通过接口使用，不直接依赖本包细节。
type Service struct {
	mu              sync.RWMutex
	downgradeSwitch string
	cutRange        string
	scBlacklist     string
	cacheSwitch     string
}

func New(downgrade, cutRange, scBlacklist, cacheSwitch string) *Service {
	return &Service{
		downgradeSwitch: downgrade,
		cutRange:        cutRange,
		scBlacklist:     scBlacklist,
		cacheSwitch:     cacheSwitch,
	}
}

func (s *Service) IsDowngradeSwitch() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.downgradeSwitch == "1"
}

// IsCutRange 用户切量：hash 后两位 <= cutRange
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

// IsSCBlackIntercept true=拦截
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

// Update 运行时热更新（DCC HTTP 接口调用）
func (s *Service) Update(key, value string) {
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
	}
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

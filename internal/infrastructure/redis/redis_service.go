package redis

import (
	"context"
	"hash/fnv"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Service Redis 封装（库存占用、人群 BitSet、缓存）
type Service struct {
	client *goredis.Client
}

func New(addr, password string, db int) *Service {
	return &Service{
		client: goredis.NewClient(&goredis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		}),
	}
}

func (s *Service) Client() *goredis.Client { return s.client }

func (s *Service) Ping(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}

func (s *Service) Incr(ctx context.Context, key string) (int64, error) {
	return s.client.Incr(ctx, key).Result()
}

func (s *Service) GetInt64(ctx context.Context, key string) (int64, error) {
	n, err := s.client.Get(ctx, key).Int64()
	if err == goredis.Nil {
		return 0, nil
	}
	return n, err
}

func (s *Service) SetNX(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return s.client.SetNX(ctx, key, "1", expiration).Result()
}

func (s *Service) Del(ctx context.Context, keys ...string) error {
	return s.client.Del(ctx, keys...).Err()
}

// IndexFromUserID 将 userId 映射为 bit 下标（对齐 Java getIndexFromUserId）
func (s *Service) IndexFromUserID(userID string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(userID))
	return int64(h.Sum64() % 100000000)
}

// IsTagCrowdRange BitSet 人群判断；key 不存在则放行 true
func (s *Service) IsTagCrowdRange(ctx context.Context, tagID, userID string) (bool, error) {
	exists, err := s.client.Exists(ctx, tagID).Result()
	if err != nil {
		return false, err
	}
	if exists == 0 {
		return true, nil
	}
	idx := s.IndexFromUserID(userID)
	bit, err := s.client.GetBit(ctx, tagID, idx).Result()
	if err != nil {
		return false, err
	}
	return bit == 1, nil
}

// AddTagCrowd 写入人群标签 bit
func (s *Service) AddTagCrowd(ctx context.Context, tagID, userID string) error {
	idx := s.IndexFromUserID(userID)
	return s.client.SetBit(ctx, tagID, idx, 1).Err()
}

// TryLock 简易分布式锁 SET NX EX；成功返回 true
func (s *Service) TryLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return s.client.SetNX(ctx, key, "1", ttl).Result()
}

// Unlock 释放锁
func (s *Service) Unlock(ctx context.Context, key string) error {
	return s.client.Del(ctx, key).Err()
}

// TryLockWait 带等待的抢锁（对齐 Redisson tryLock wait/lease）
func (s *Service) TryLockWait(ctx context.Context, key string, wait, lease time.Duration) (bool, error) {
	deadline := time.Now().Add(wait)
	for {
		ok, err := s.TryLock(ctx, key, lease)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
		if time.Now().After(deadline) {
			return false, nil
		}
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-time.After(50 * time.Millisecond):
		}
	}
}

func (s *Service) Close() error {
	return s.client.Close()
}

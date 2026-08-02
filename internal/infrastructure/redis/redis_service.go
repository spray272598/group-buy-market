package redis

import (
	"context"
	"hash/fnv"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"group-buy-market/internal/infrastructure/metrics"
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

func observe(op string, start time.Time, err error) {
	metrics.ObserveRedis(op, err, time.Since(start).Seconds())
}

func (s *Service) Ping(ctx context.Context) error {
	start := time.Now()
	err := s.client.Ping(ctx).Err()
	observe("ping", start, err)
	return err
}

func (s *Service) Incr(ctx context.Context, key string) (int64, error) {
	start := time.Now()
	n, err := s.client.Incr(ctx, key).Result()
	observe("incr", start, err)
	return n, err
}

func (s *Service) GetInt64(ctx context.Context, key string) (int64, error) {
	start := time.Now()
	n, err := s.client.Get(ctx, key).Int64()
	if err == goredis.Nil {
		observe("get", start, nil)
		return 0, nil
	}
	observe("get", start, err)
	return n, err
}

func (s *Service) SetNX(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	start := time.Now()
	ok, err := s.client.SetNX(ctx, key, "1", expiration).Result()
	observe("setnx", start, err)
	return ok, err
}

func (s *Service) Del(ctx context.Context, keys ...string) error {
	start := time.Now()
	err := s.client.Del(ctx, keys...).Err()
	observe("del", start, err)
	return err
}

// IndexFromUserID 将 userId 映射为 bit 下标（对齐 Java getIndexFromUserId）
func (s *Service) IndexFromUserID(userID string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(userID))
	return int64(h.Sum64() % 100000000)
}

// IsTagCrowdRange BitSet 人群判断；key 不存在则放行 true
func (s *Service) IsTagCrowdRange(ctx context.Context, tagID, userID string) (bool, error) {
	start := time.Now()
	exists, err := s.client.Exists(ctx, tagID).Result()
	if err != nil {
		observe("exists", start, err)
		return false, err
	}
	if exists == 0 {
		observe("getbit", start, nil)
		return true, nil
	}
	idx := s.IndexFromUserID(userID)
	bit, err := s.client.GetBit(ctx, tagID, idx).Result()
	observe("getbit", start, err)
	if err != nil {
		return false, err
	}
	return bit == 1, nil
}

// AddTagCrowd 写入人群标签 bit
func (s *Service) AddTagCrowd(ctx context.Context, tagID, userID string) error {
	start := time.Now()
	idx := s.IndexFromUserID(userID)
	err := s.client.SetBit(ctx, tagID, idx, 1).Err()
	observe("setbit", start, err)
	return err
}

// TryLock 简易分布式锁 SET NX EX；成功返回 true
func (s *Service) TryLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return s.SetNX(ctx, key, ttl)
}

// Unlock 释放锁
func (s *Service) Unlock(ctx context.Context, key string) error {
	return s.Del(ctx, key)
}

// TryLockWait 带等待的抢锁（对齐 Redisson tryLock wait/lease）
func (s *Service) TryLockWait(ctx context.Context, key string, wait, lease time.Duration) (bool, error) {
	start := time.Now()
	deadline := time.Now().Add(wait)
	for {
		ok, err := s.TryLock(ctx, key, lease)
		if err != nil {
			observe("try_lock_wait", start, err)
			return false, err
		}
		if ok {
			observe("try_lock_wait", start, nil)
			return true, nil
		}
		if time.Now().After(deadline) {
			observe("try_lock_wait", start, nil)
			return false, nil
		}
		select {
		case <-ctx.Done():
			observe("try_lock_wait", start, ctx.Err())
			return false, ctx.Err()
		case <-time.After(50 * time.Millisecond):
		}
	}
}

// Publish Redis Pub/Sub 发布（DCC 跨实例）
func (s *Service) Publish(ctx context.Context, channel string, message string) error {
	start := time.Now()
	err := s.client.Publish(ctx, channel, message).Err()
	observe("publish", start, err)
	return err
}

// Subscribe Redis Pub/Sub 订阅，handler 在独立 goroutine 中串行调用
func (s *Service) Subscribe(ctx context.Context, channel string, handler func(payload string)) error {
	start := time.Now()
	pubsub := s.client.Subscribe(ctx, channel)
	// 等待订阅确认
	if _, err := pubsub.Receive(ctx); err != nil {
		_ = pubsub.Close()
		observe("subscribe", start, err)
		return err
	}
	observe("subscribe", start, nil)
	ch := pubsub.Channel()
	go func() {
		defer pubsub.Close()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				if msg != nil {
					handler(msg.Payload)
				}
			}
		}
	}()
	return nil
}

func (s *Service) Close() error {
	return s.client.Close()
}

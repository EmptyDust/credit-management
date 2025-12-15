package utils

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(addr, password string, db int) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
		PoolSize: 10,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("Redis连接失败: %v", err)
		return nil
	}

	log.Println("Redis连接成功")
	return &RedisClient{client: client}
}

// AddToBlacklist 将token添加到黑名单
func (r *RedisClient) AddToBlacklist(ctx context.Context, token string, expiration time.Duration) error {
	key := fmt.Sprintf("blacklist:%s", token)
	return r.client.Set(ctx, key, "revoked", expiration).Err()
}

// IsBlacklisted 检查token是否在黑名单中
func (r *RedisClient) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	key := fmt.Sprintf("blacklist:%s", token)
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// SetUserSession 设置用户会话信息
func (r *RedisClient) SetUserSession(ctx context.Context, userID string, sessionData map[string]interface{}, expiration time.Duration) error {
	key := fmt.Sprintf("session:%s", userID)
	return r.client.HMSet(ctx, key, sessionData).Err()
}

// GetUserSession 获取用户会话信息
func (r *RedisClient) GetUserSession(ctx context.Context, userID string) (map[string]string, error) {
	key := fmt.Sprintf("session:%s", userID)
	return r.client.HGetAll(ctx, key).Result()
}

// DeleteUserSession 删除用户会话
func (r *RedisClient) DeleteUserSession(ctx context.Context, userID string) error {
	key := fmt.Sprintf("session:%s", userID)
	return r.client.Del(ctx, key).Err()
}

// SetCache 设置缓存
func (r *RedisClient) SetCache(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// GetCache 获取缓存
func (r *RedisClient) GetCache(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// DeleteCache 删除缓存
func (r *RedisClient) DeleteCache(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Close 关闭Redis连接
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// IncrementRateLimit 增加速率限制计数器
// 返回当前计数和是否超过限制
func (r *RedisClient) IncrementRateLimit(ctx context.Context, key string, limit int64, window time.Duration) (int64, bool, error) {
	// 增加计数
	count, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, false, err
	}

	// 如果是第一次访问，设置过期时间
	if count == 1 {
		if err := r.client.Expire(ctx, key, window).Err(); err != nil {
			return count, false, err
		}
	}

	// 检查是否超过限制
	exceeded := count > limit
	return count, exceeded, nil
}

// GetRateLimitCount 获取当前速率限制计数
func (r *RedisClient) GetRateLimitCount(ctx context.Context, key string) (int64, error) {
	count, err := r.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return count, err
}


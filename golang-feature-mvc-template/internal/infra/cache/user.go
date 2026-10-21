// Package cache 基于 Redis 的缓存实现,适配 feature service 定义的
// Cache 端口(Get/Set/Del)。接口定义在消费方,本包只提供具体实现。
package cache

import (
	"context"
	"time"

	commoncache "github.com/Choi-Keith/golang-common/cache"
)

// Cache 包装 golang-common/cache 的 RedisCache,屏蔽 go-redis 细节。
type Cache struct {
	c *commoncache.RedisCache
}

// New 创建 Redis 缓存。
func New(c *commoncache.RedisCache) *Cache {
	return &Cache{c: c}
}

// Get 读取 key,key 不存在时返回错误(调用方按 cache miss 处理)。
func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	return c.c.Client().Get(ctx, key).Result()
}

// Set 写入 key 并设置过期时间。
func (c *Cache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return c.c.Client().Set(ctx, key, value, ttl).Err()
}

// Del 删除若干 key,删除不存在的 key 不报错。
func (c *Cache) Del(ctx context.Context, keys ...string) error {
	return c.c.Client().Del(ctx, keys...).Err()
}

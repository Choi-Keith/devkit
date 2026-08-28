// Package container 依赖注入容器:连接基础设施、迁移模型、组装各层组件。
package container

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/Choi-Keith/golang-common/cache"
	"github.com/Choi-Keith/golang-common/database"
	"github.com/Choi-Keith/golang-common/health"

	"golang-simple-template/internal/config"
	"golang-simple-template/internal/handler"
	"golang-simple-template/internal/model"
	"golang-simple-template/internal/repository"
	"golang-simple-template/internal/service"
)

// Container 依赖注入容器
type Container struct {
	Config *config.Config
	DB     *gorm.DB
	Cache  *cache.RedisCache

	// Repositories
	UserRepo repository.UserRepository

	// Services
	UserService service.UserService

	// Handlers
	HealthHandler *health.HealthHandler
	UserHandler   *handler.UserHandler
}

// New 初始化容器及所有依赖
func New(cfg *config.Config) (*Container, error) {
	c := &Container{Config: cfg}

	// 数据库
	db, err := database.New(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("init database: %w", err)
	}
	c.DB = db

	// 自动迁移模型
	if err := db.AutoMigrate(&model.User{}); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}

	// Redis 缓存
	redisCache, err := cache.NewRedisCache(cfg.Redis)
	if err != nil {
		return nil, fmt.Errorf("init redis: %w", err)
	}
	c.Cache = redisCache

	// Repositories
	c.UserRepo = repository.NewUserRepository(db)

	// Services
	c.UserService = service.NewUserService(c.UserRepo)

	// Handlers
	c.HealthHandler = health.NewHealthHandler("1.0.0")
	c.HealthHandler.AddCheckers(map[string]health.Checker{
		"database": database.HealthChecker(db),
		"redis": health.NewChecker(func(ctx context.Context) error {
			return redisCache.Client().Ping(ctx).Err()
		}),
	})
	c.UserHandler = handler.NewUserHandler(c.UserService)

	return c, nil
}

// Close 释放容器持有的资源
func (c *Container) Close() {
	if c.Cache != nil {
		_ = c.Cache.Close()
	}
	if c.DB != nil {
		_ = database.Close(c.DB)
	}
}

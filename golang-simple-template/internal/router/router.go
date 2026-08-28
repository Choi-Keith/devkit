// Package router 路由注册:全局中间件与路由分组。
package router

import (
	"github.com/gin-gonic/gin"

	"github.com/Choi-Keith/golang-common/middleware"

	"golang-simple-template/internal/config"
	"golang-simple-template/internal/container"
)

// Setup 初始化路由
func Setup(cfg *config.Config, c *container.Container) *gin.Engine {
	if cfg.Server.Production {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// 全局中间件
	r.Use(middleware.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logging(middleware.LoggingConfig{
		SkipPaths: []string{"/health", "/health/live", "/health/ready"},
	}))
	r.Use(middleware.CORS())

	// 健康检查:/health, /health/live, /health/ready
	c.HealthHandler.RegisterRoutes(r)

	// API v1 路由
	v1 := r.Group("/api/v1")
	{
		users := v1.Group("/users")
		{
			users.POST("", c.UserHandler.Create)
			users.GET("", c.UserHandler.List)
			users.GET("/:id", c.UserHandler.GetByID)
			users.PUT("/:id", c.UserHandler.Update)
			users.DELETE("/:id", c.UserHandler.Delete)
		}
	}

	return r
}

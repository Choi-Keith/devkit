// Package user 用户 feature 根包:只承载路由注册(router.go)。
// 分层见子包:handler/(接口层 + ecode)、service/(应用层,定义 Repository/Cache 端口)、
// dto/(请求/响应 DTO);基础设施实现位于 internal/infra/ 并由 main 装配注入。
package user

import (
	"github.com/gin-gonic/gin"

	"github.com/Choi-Keith/golang-feature-mvc-template/internal/user/handler"
)

// RegisterRoutes 把用户模块的路由注册到 /api/v1 分组下。
// 注册/登录/刷新/登出公开;CRUD 接口按需自行追加认证中间件。
func RegisterRoutes(rg *gin.RouterGroup, h *handler.Handler) {
	users := rg.Group("/users")
	users.POST("/register", h.Register)
	users.POST("/login", h.Login)
	users.POST("/refresh", h.Refresh)
	users.POST("/logout", h.Logout)

	users.GET("", h.List)
	users.GET("/:id", h.GetByID)
	users.PUT("/:id", h.Update)
	users.DELETE("/:id", h.Delete)
}

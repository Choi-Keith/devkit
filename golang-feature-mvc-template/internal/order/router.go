// Package order 订单 feature 根包:只承载路由注册(router.go)。
// 分层见子包:handler/(接口层 + ecode)、service/(应用层,定义 Repository 端口)、
// dto/(请求/响应 DTO);基础设施实现位于 internal/infra/ 并由 main 装配注入。
package order

import (
	"github.com/gin-gonic/gin"

	"github.com/Choi-Keith/golang-feature-mvc-template/internal/order/handler"
)

// RegisterRoutes 把订单模块的路由注册到 /api/v1 分组下,
// 所有订单接口都要求 JWT 认证。
func RegisterRoutes(rg *gin.RouterGroup, h *handler.Handler, jwtMiddleware gin.HandlerFunc) {
	orders := rg.Group("/orders", jwtMiddleware)
	orders.POST("", h.Create)
	orders.GET("", h.List)
	orders.GET("/:id", h.GetByID)
	orders.POST("/:id/cancel", h.Cancel)
}

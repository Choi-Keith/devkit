// Package dto 订单模块的请求/响应 DTO。统一包络与分页类型见
// pkg/dto(Resp[T] / ListResp[T] / ErrorResp / PageReq)。
package dto

import (
	pkgdto "github.com/Choi-Keith/golang-feature-mvc-template/pkg/dto"
)

// CreateOrderReq 是 POST /api/v1/orders 的请求体。
type CreateOrderReq struct {
	UserID uint64 `json:"userId" binding:"required" example:"1"`            // 下单用户 ID
	Amount int64  `json:"amount" binding:"required,min=1" example:"199900"` // 金额,单位:分
}

// ListOrderReq 是 GET /api/v1/orders 的查询参数。
type ListOrderReq struct {
	pkgdto.PageReq
}

// OrderResp 是订单的 HTTP 视图。
type OrderResp struct {
	ID        uint64 `json:"id" example:"1"`
	UserID    uint64 `json:"userId" example:"1"`
	OrderNo   string `json:"orderNo" example:"O17248032000000000001234"`
	Amount    int64  `json:"amount" example:"199900"`
	Status    string `json:"status" example:"pending"`
	CreatedAt string `json:"createdAt" example:"2026-08-27T08:00:00Z"`
}

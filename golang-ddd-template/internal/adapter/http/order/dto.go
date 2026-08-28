package orderhttp

// CreateOrderRequest is the payload of POST /api/v1/orders.
type CreateOrderRequest struct {
	UserID uint64 `json:"userId" binding:"required,min=1" example:"1"`      // 下单用户 ID
	Amount int64  `json:"amount" binding:"required,min=1" example:"199900"` // 金额,单位:分
}

// ListOrderRequest is the query of GET /api/v1/orders.
type ListOrderRequest struct {
	Page     int `form:"page" binding:"omitempty,min=1" example:"1"`              // 页码,从 1 开始
	PageSize int `form:"pageSize" binding:"omitempty,min=1,max=100" example:"20"` // 每页数量,最大 100
}

// OrderResponse is the HTTP view of an order.
type OrderResponse struct {
	ID        uint64 `json:"id" example:"1"`
	UserID    uint64 `json:"userId" example:"1"`
	OrderNo   string `json:"orderNo" example:"O17562624000000000000123"`
	Amount    int64  `json:"amount" example:"199900"`
	Status    string `json:"status" example:"pending"`
	CreatedAt string `json:"createdAt" example:"2026-08-27T08:00:00Z"`
}

// OrderListData is the data payload of the order list endpoint.
type OrderListData struct {
	List     []OrderResponse `json:"list"`
	Total    int64           `json:"total" example:"100"`
	Page     int             `json:"page" example:"1"`
	PageSize int             `json:"pageSize" example:"20"`
}

// Swagger documentation envelopes. The runtime envelope is provided by
// golang-common/response; these mirror it so swag can render concrete schemas.

// ErrorEnvelope documents an error response.
type ErrorEnvelope struct {
	Code    int    `json:"code" example:"20100"`
	Message string `json:"message" example:"订单不存在"`
}

// OrderEnvelope documents a single-order response.
type OrderEnvelope struct {
	Code    int           `json:"code" example:"0"`
	Message string        `json:"message" example:"success"`
	Data    OrderResponse `json:"data"`
}

// OrderListEnvelope documents an order-list response.
type OrderListEnvelope struct {
	Code    int           `json:"code" example:"0"`
	Message string        `json:"message" example:"success"`
	Data    OrderListData `json:"data"`
}

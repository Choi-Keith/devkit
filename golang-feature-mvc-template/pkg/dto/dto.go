// Package dto 全局共享的请求/响应类型:统一包络的泛型定义与分页类型。
// 运行时的响应包络由 golang-common/response 提供,这里的泛型类型镜像其
// JSON 结构,供 handler 的 swagger 注解渲染出具体的 schema。
package dto

// Resp 成功包络,镜像 golang-common/response.Response。
type Resp[T any] struct {
	Code      int    `json:"code" example:"0"`                                         // 业务码,0 为成功
	Message   string `json:"message" example:"success"`                                // 响应消息
	Data      T      `json:"data"`                                                     // 响应数据
	RequestID string `json:"requestId" example:"550e8400-e29b-41d4-a716-446655440000"` // 请求 ID,链路追踪用
}

// ErrorResp 错误包络,全局唯一一份,不带 feature 前缀。
type ErrorResp struct {
	Code    int    `json:"code" example:"20000"`
	Message string `json:"message" example:"用户不存在"`
}

// ListData 分页负载,镜像 golang-common/response.PageData。
type ListData[T any] struct {
	List     []T   `json:"list"`
	Total    int64 `json:"total" example:"100"`
	Page     int   `json:"page" example:"1"`
	PageSize int   `json:"pageSize" example:"20"`
}

// ListResp 分页包络,Data 为 ListData[T]。
type ListResp[T any] struct {
	Code      int         `json:"code" example:"0"`
	Message   string      `json:"message" example:"success"`
	Data      ListData[T] `json:"data"`
	RequestID string      `json:"requestId" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// PageReq 分页查询参数,供各 feature 的 List 请求内嵌。
type PageReq struct {
	Page     int `form:"page" binding:"omitempty,min=1" example:"1"`              // 页码,从 1 开始
	PageSize int `form:"pageSize" binding:"omitempty,min=1,max=100" example:"20"` // 每页数量,最大 100
}

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

// NormalizePage 分页参数归一化:page 最小 1,pageSize 默认 20、上限 100。
// handler(响应回显)与 service(查询)共用这一份实现。
func NormalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return page, pageSize
}

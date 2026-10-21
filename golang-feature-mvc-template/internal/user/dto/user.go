// Package dto 用户模块的请求/响应 DTO。统一包络与分页类型见
// pkg/dto(Resp[T] / ListResp[T] / ErrorResp / PageReq)。
package dto

import (
	pkgdto "github.com/Choi-Keith/golang-feature-mvc-template/pkg/dto"
)

// RegisterReq 是 POST /api/v1/users/register 的请求体。
type RegisterReq struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"` // 邮箱
	Name     string `json:"name" binding:"required" example:"Alice"`                   // 昵称
	Password string `json:"password" binding:"required,min=8" example:"P@ssw0rd123"`   // 密码(至少 8 位)
}

// LoginReq 是 POST /api/v1/users/login 的请求体。
type LoginReq struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"` // 邮箱
	Password string `json:"password" binding:"required" example:"P@ssw0rd123"`         // 密码
}

// RefreshReq 是 POST /api/v1/users/refresh 的请求体。
type RefreshReq struct {
	RefreshToken string `json:"refreshToken" binding:"required"` // 登录时签发的 refresh token
}

// LogoutReq 是 POST /api/v1/users/logout 的请求体。
type LogoutReq struct {
	RefreshToken string `json:"refreshToken" binding:"required"` // 要吊销的 refresh token
}

// UpdateUserReq 是 PUT /api/v1/users/:id 的请求体,字段均可选。
type UpdateUserReq struct {
	Name     string `json:"name" binding:"omitempty" example:"Alice"`                 // 新昵称
	Password string `json:"password" binding:"omitempty,min=8" example:"N3wP@ssw0rd"` // 新密码(至少 8 位)
}

// ListUserReq 是 GET /api/v1/users 的查询参数。
type ListUserReq struct {
	pkgdto.PageReq
}

// UserResp 是用户的 HTTP 视图,不含密码。
type UserResp struct {
	ID        uint64 `json:"id" example:"1"`
	Email     string `json:"email" example:"user@example.com"`
	Name      string `json:"name" example:"Alice"`
	CreatedAt string `json:"createdAt" example:"2026-08-27T08:00:00Z"`
}

// TokenResp 登录/刷新成功后返回的令牌对。
type TokenResp struct {
	AccessToken  string `json:"accessToken"`  // 访问令牌,2 小时有效
	RefreshToken string `json:"refreshToken"` // 刷新令牌,7 天有效
	TokenType    string `json:"tokenType" example:"Bearer"`
}

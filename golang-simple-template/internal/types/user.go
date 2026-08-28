// Package types 定义请求/响应 DTO,与 model 层分离。
// JSON 字段统一使用 camelCase。
package types

import "time"

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Email string `json:"email" binding:"required,email"`
	Name  string `json:"name" binding:"required,max=64"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Name string `json:"name" binding:"required,max=64"`
}

// UserResponse 用户响应
type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

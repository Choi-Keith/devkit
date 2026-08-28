package userhttp

// RegisterUserRequest is the payload of POST /api/v1/users.
type RegisterUserRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"` // 邮箱
	Name     string `json:"name" binding:"required" example:"Alice"`                   // 昵称
	Password string `json:"password" binding:"required,min=8" example:"P@ssw0rd123"`   // 密码(至少 8 位)
}

// ListUserRequest is the query of GET /api/v1/users.
type ListUserRequest struct {
	Page     int `form:"page" binding:"omitempty,min=1" example:"1"`              // 页码,从 1 开始
	PageSize int `form:"pageSize" binding:"omitempty,min=1,max=100" example:"20"` // 每页数量,最大 100
}

// UserResponse is the HTTP view of a user.
type UserResponse struct {
	ID        uint64 `json:"id" example:"1"`
	Email     string `json:"email" example:"user@example.com"`
	Name      string `json:"name" example:"Alice"`
	CreatedAt string `json:"createdAt" example:"2026-08-27T08:00:00Z"`
}

// UserListData is the data payload of the user list endpoint.
type UserListData struct {
	List     []UserResponse `json:"list"`
	Total    int64          `json:"total" example:"100"`
	Page     int            `json:"page" example:"1"`
	PageSize int            `json:"pageSize" example:"20"`
}

// Swagger documentation envelopes. The runtime envelope is provided by
// golang-common/response; these mirror it so swag can render concrete schemas.

// ErrorEnvelope documents an error response.
type ErrorEnvelope struct {
	Code    int    `json:"code" example:"20000"`
	Message string `json:"message" example:"用户不存在"`
}

// UserEnvelope documents a single-user response.
type UserEnvelope struct {
	Code    int          `json:"code" example:"0"`
	Message string       `json:"message" example:"success"`
	Data    UserResponse `json:"data"`
}

// UserListEnvelope documents a user-list response.
type UserListEnvelope struct {
	Code    int          `json:"code" example:"0"`
	Message string       `json:"message" example:"success"`
	Data    UserListData `json:"data"`
}

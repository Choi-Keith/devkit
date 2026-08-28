// Package handler HTTP 处理器:参数校验、调用 service、统一响应。
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/Choi-Keith/golang-common/response"

	"golang-simple-template/internal/ecode"
	"golang-simple-template/internal/service"
	"golang-simple-template/internal/types"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// Create 创建用户
// POST /api/v1/users
func (h *UserHandler) Create(c *gin.Context) {
	var req types.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, ecode.ErrInvalidParams.WithDetails(err.Error()))
		return
	}

	user, err := h.userService.Create(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, user)
}

// GetByID 获取用户详情
// GET /api/v1/users/:id
func (h *UserHandler) GetByID(c *gin.Context) {
	user, err := h.userService.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, user)
}

// List 用户列表(分页)
// GET /api/v1/users?page=1&pageSize=20
func (h *UserHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	users, total, err := h.userService.List(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, users, total, page, pageSize)
}

// Update 更新用户
// PUT /api/v1/users/:id
func (h *UserHandler) Update(c *gin.Context) {
	var req types.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, ecode.ErrInvalidParams.WithDetails(err.Error()))
		return
	}

	user, err := h.userService.Update(c.Request.Context(), c.Param("id"), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, user)
}

// Delete 删除用户
// DELETE /api/v1/users/:id
func (h *UserHandler) Delete(c *gin.Context) {
	if err := h.userService.Delete(c.Request.Context(), c.Param("id")); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

package userhttp

import (
	"context"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/Choi-Keith/golang-common/ecode"
	"github.com/Choi-Keith/golang-common/response"

	"github.com/Choi-Keith/golang-ddd-template/internal/application/user"
	"github.com/Choi-Keith/golang-ddd-template/internal/domain/user"
)

// UserUseCase is the slice of the application service the handler needs.
// Defined here so the handler stays testable with a fake implementation.
type UserUseCase interface {
	Register(ctx context.Context, in userapp.RegisterInput) (*userapp.UserDTO, error)
	GetByID(ctx context.Context, id uint64) (*userapp.UserDTO, error)
	List(ctx context.Context, page, pageSize int) (*userapp.UserListDTO, error)
	Delete(ctx context.Context, id uint64) error
}

// UserHandler exposes the user use cases over HTTP.
type UserHandler struct {
	svc UserUseCase
}

// NewUserHandler creates a UserHandler.
func NewUserHandler(svc UserUseCase) *UserHandler {
	return &UserHandler{svc: svc}
}

// Register godoc
//
//	@Summary		注册用户
//	@Description	创建一个新用户账号
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			request	body		RegisterUserRequest	true	"注册请求"
//	@Success		201		{object}	UserEnvelope		"创建成功"
//	@Failure		400		{object}	ErrorEnvelope		"参数无效"
//	@Failure		409		{object}	ErrorEnvelope		"邮箱已被注册"
//	@Router			/users [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, ecode.ErrInvalidParams.WithDetails(err.Error()))
		return
	}

	out, err := h.svc.Register(c.Request.Context(), userapp.RegisterInput{
		Email:    req.Email,
		Name:     req.Name,
		Password: req.Password,
	})
	if err != nil {
		handleError(c, err)
		return
	}
	response.Created(c, toUserResponse(out))
}

// GetByID godoc
//
//	@Summary		查询用户详情
//	@Tags			user
//	@Produce		json
//	@Param			id	path		integer			true	"用户 ID"	example(1)
//	@Success		200	{object}	UserEnvelope	"查询成功"
//	@Failure		400	{object}	ErrorEnvelope	"参数无效"
//	@Failure		404	{object}	ErrorEnvelope	"用户不存在"
//	@Router			/users/{id} [get]
func (h *UserHandler) GetByID(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, ecode.ErrInvalidParams.WithDetails("id must be a positive integer"))
		return
	}
	out, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, toUserResponse(out))
}

// List godoc
//
//	@Summary		分页查询用户
//	@Tags			user
//	@Produce		json
//	@Param			page		query		int	false	"页码,从 1 开始"		example(1)
//	@Param			pageSize	query		int	false	"每页数量,最大 100"	example(20)
//	@Success		200			{object}	UserListEnvelope	"查询成功"
//	@Failure		400			{object}	ErrorEnvelope		"参数无效"
//	@Router			/users [get]
func (h *UserHandler) List(c *gin.Context) {
	var req ListUserRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, ecode.ErrInvalidParams.WithDetails(err.Error()))
		return
	}

	out, err := h.svc.List(c.Request.Context(), req.Page, req.PageSize)
	if err != nil {
		handleError(c, err)
		return
	}

	list := make([]UserResponse, 0, len(out.List))
	for i := range out.List {
		list = append(list, toUserResponse(&out.List[i]))
	}
	response.Page(c, list, out.Total, out.Page, out.PageSize)
}

// Delete godoc
//
//	@Summary		删除用户
//	@Tags			user
//	@Produce		json
//	@Param			id	path		integer	true	"用户 ID"	example(1)
//	@Success		204	"删除成功"
//	@Failure		400	{object}	ErrorEnvelope	"参数无效"
//	@Failure		404	{object}	ErrorEnvelope	"用户不存在"
//	@Router			/users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, ecode.ErrInvalidParams.WithDetails("id must be a positive integer"))
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		handleError(c, err)
		return
	}
	response.NoContent(c)
}

// parseID parses the :id path parameter as a bigint unsigned id.
func parseID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("id"), 10, 64)
}

// handleError maps domain errors to business error codes at the boundary.
func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, user.ErrNotFound):
		response.Error(c, ErrUserNotFound)
	case errors.Is(err, user.ErrEmailTaken):
		response.Error(c, ErrEmailTaken)
	case errors.Is(err, user.ErrEmailRequired),
		errors.Is(err, user.ErrInvalidEmail),
		errors.Is(err, user.ErrNameRequired),
		errors.Is(err, user.ErrPasswordRequired):
		response.Error(c, ecode.ErrInvalidParams.WithDetails(err.Error()))
	default:
		response.Error(c, ecode.ErrInternalServer.WithDetails(err.Error()))
	}
}

func toUserResponse(d *userapp.UserDTO) UserResponse {
	return UserResponse{
		ID:        d.ID,
		Email:     d.Email,
		Name:      d.Name,
		CreatedAt: d.CreatedAt,
	}
}

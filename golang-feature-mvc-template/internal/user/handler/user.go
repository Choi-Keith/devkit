// Package handler 用户模块接口层:HTTP 绑定与 swagger 注解。
// service 返回的 pkg/xerr 业务错误码直接交给 response.Error
// (内部 ecode.FromError 识别 *ecode.Error 并保留 code)。
package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Choi-Keith/golang-common/ecode"
	"github.com/Choi-Keith/golang-common/response"

	"github.com/Choi-Keith/golang-feature-mvc-template/internal/infra/db/user/model"
	"github.com/Choi-Keith/golang-feature-mvc-template/internal/user/dto"
	"github.com/Choi-Keith/golang-feature-mvc-template/internal/user/service"
	pkgdto "github.com/Choi-Keith/golang-feature-mvc-template/pkg/dto"
)

// Handler 暴露用户模块的 HTTP 接口。
type Handler struct {
	svc *service.Service
}

// NewHandler 创建用户 Handler。
func NewHandler(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

// Register godoc
//
//	@Summary		注册用户
//	@Description	创建一个新用户账号
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.RegisterReq			true	"注册请求"
//	@Success		201		{object}	pkgdto.Resp[dto.UserResp]	"创建成功"
//	@Failure		400		{object}	pkgdto.ErrorResp			"参数无效"
//	@Failure		409		{object}	pkgdto.ErrorResp			"邮箱已被注册"
//	@Router			/users/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req dto.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, ecode.ErrInvalidParams.WithDetails(err.Error()))
		return
	}

	u, err := h.svc.Register(c.Request.Context(), req.Email, req.Name, req.Password)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, toUserResp(u))
}

// Login godoc
//
//	@Summary		登录
//	@Description	校验邮箱密码,签发 access token(2 小时)与 refresh token(7 天)
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.LoginReq				true	"登录请求"
//	@Success		200		{object}	pkgdto.Resp[dto.TokenResp]	"登录成功"
//	@Failure		400		{object}	pkgdto.ErrorResp			"参数无效"
//	@Failure		401		{object}	pkgdto.ErrorResp			"邮箱或密码错误"
//	@Router			/users/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req dto.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, ecode.ErrInvalidParams.WithDetails(err.Error()))
		return
	}

	access, refresh, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, toTokenResp(access, refresh))
}

// Refresh godoc
//
//	@Summary		刷新令牌
//	@Description	用 refresh token 换发新的令牌对,旧 refresh token 立即失效
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.RefreshReq				true	"刷新请求"
//	@Success		200		{object}	pkgdto.Resp[dto.TokenResp]	"刷新成功"
//	@Failure		400		{object}	pkgdto.ErrorResp			"参数无效"
//	@Failure		401		{object}	pkgdto.ErrorResp			"refresh token 无效或已吊销"
//	@Router			/users/refresh [post]
func (h *Handler) Refresh(c *gin.Context) {
	var req dto.RefreshReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, ecode.ErrInvalidParams.WithDetails(err.Error()))
		return
	}

	access, refresh, err := h.svc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, toTokenResp(access, refresh))
}

// Logout godoc
//
//	@Summary		登出
//	@Description	吊销 refresh token(删除 Redis 中的哈希),幂等
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.LogoutReq		true	"登出请求"
//	@Success		204		"登出成功"
//	@Failure		400		{object}	pkgdto.ErrorResp	"参数无效"
//	@Failure		401		{object}	pkgdto.ErrorResp	"refresh token 无效"
//	@Router			/users/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	var req dto.LogoutReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, ecode.ErrInvalidParams.WithDetails(err.Error()))
		return
	}
	if err := h.svc.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

// GetByID godoc
//
//	@Summary		查询用户详情
//	@Description	cache-aside:先查 Redis,miss 走 DB 并回写
//	@Tags			user
//	@Produce		json
//	@Param			id	path		integer						true	"用户 ID"	example(1)
//	@Success		200	{object}	pkgdto.Resp[dto.UserResp]	"查询成功"
//	@Failure		400	{object}	pkgdto.ErrorResp			"参数无效"
//	@Failure		404	{object}	pkgdto.ErrorResp			"用户不存在"
//	@Router			/users/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, ecode.ErrInvalidParams.WithDetails("id must be a positive integer"))
		return
	}
	u, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, toUserResp(u))
}

// List godoc
//
//	@Summary		分页查询用户
//	@Tags			user
//	@Produce		json
//	@Param			page		query		int	false	"页码,从 1 开始"		example(1)
//	@Param			pageSize	query		int	false	"每页数量,最大 100"	example(20)
//	@Success		200			{object}	pkgdto.ListResp[dto.UserResp]	"查询成功"
//	@Failure		400			{object}	pkgdto.ErrorResp				"参数无效"
//	@Router			/users [get]
func (h *Handler) List(c *gin.Context) {
	var req dto.ListUserReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, ecode.ErrInvalidParams.WithDetails(err.Error()))
		return
	}

	page, pageSize := pkgdto.NormalizePage(req.Page, req.PageSize)
	users, total, err := h.svc.List(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}

	list := make([]dto.UserResp, 0, len(users))
	for _, u := range users {
		list = append(list, toUserResp(u))
	}
	response.Page(c, list, total, page, pageSize)
}

// Update godoc
//
//	@Summary		更新用户
//	@Description	修改昵称,可一并重置密码;成功后使详情缓存失效
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			id		path		integer						true	"用户 ID"	example(1)
//	@Param			request	body		dto.UpdateUserReq			true	"更新请求"
//	@Success		200		{object}	pkgdto.Resp[dto.UserResp]	"更新成功"
//	@Failure		400		{object}	pkgdto.ErrorResp			"参数无效"
//	@Failure		404		{object}	pkgdto.ErrorResp			"用户不存在"
//	@Router			/users/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, ecode.ErrInvalidParams.WithDetails("id must be a positive integer"))
		return
	}
	var req dto.UpdateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, ecode.ErrInvalidParams.WithDetails(err.Error()))
		return
	}

	u, err := h.svc.Update(c.Request.Context(), id, req.Name, req.Password)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, toUserResp(u))
}

// Delete godoc
//
//	@Summary		删除用户
//	@Description	删除用户并吊销其登录态
//	@Tags			user
//	@Produce		json
//	@Param			id	path		integer	true	"用户 ID"	example(1)
//	@Success		204	"删除成功"
//	@Failure		400	{object}	pkgdto.ErrorResp	"参数无效"
//	@Failure		404	{object}	pkgdto.ErrorResp	"用户不存在"
//	@Router			/users/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, ecode.ErrInvalidParams.WithDetails("id must be a positive integer"))
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

// parseID 解析 :id 路径参数为无符号整数。
func parseID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("id"), 10, 64)
}

func toUserResp(u *model.User) dto.UserResp {
	return dto.UserResp{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
	}
}

func toTokenResp(access, refresh string) dto.TokenResp {
	return dto.TokenResp{
		AccessToken:  access,
		RefreshToken: refresh,
		TokenType:    "Bearer",
	}
}

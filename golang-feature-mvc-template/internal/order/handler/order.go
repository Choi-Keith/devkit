// Package handler 订单模块接口层:HTTP 绑定与 swagger 注解。
// service 返回的 pkg/xerr 业务错误码直接交给 response.Error
// (内部 ecode.FromError 识别 *ecode.Error 并保留 code)。
// 所有路由均挂 JWT 认证中间件。
package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Choi-Keith/golang-common/ecode"
	"github.com/Choi-Keith/golang-common/response"

	"github.com/Choi-Keith/golang-feature-mvc-template/internal/infra/db/order/model"
	"github.com/Choi-Keith/golang-feature-mvc-template/internal/order/dto"
	"github.com/Choi-Keith/golang-feature-mvc-template/internal/order/service"
	pkgdto "github.com/Choi-Keith/golang-feature-mvc-template/pkg/dto"
)

// Handler 暴露订单模块的 HTTP 接口。
type Handler struct {
	svc *service.Service
}

// NewHandler 创建订单 Handler。
func NewHandler(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

// Create godoc
//
//	@Summary		创建订单
//	@Description	为指定用户创建一个待支付订单,订单号由服务端生成
//	@Tags			order
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.CreateOrderReq			true	"创建订单请求"
//	@Success		201		{object}	pkgdto.Resp[dto.OrderResp]	"创建成功"
//	@Failure		400		{object}	pkgdto.ErrorResp			"参数无效或下单用户不存在"
//	@Failure		401		{object}	pkgdto.ErrorResp			"未认证"
//	@Router			/orders [post]
func (h *Handler) Create(c *gin.Context) {
	var req dto.CreateOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, ecode.ErrInvalidParams.WithDetails(err.Error()))
		return
	}

	o, err := h.svc.Create(c.Request.Context(), req.UserID, req.Amount)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, toOrderResp(o))
}

// GetByID godoc
//
//	@Summary		查询订单详情
//	@Tags			order
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		integer						true	"订单 ID"	example(1)
//	@Success		200	{object}	pkgdto.Resp[dto.OrderResp]	"查询成功"
//	@Failure		400	{object}	pkgdto.ErrorResp			"参数无效"
//	@Failure		401	{object}	pkgdto.ErrorResp			"未认证"
//	@Failure		404	{object}	pkgdto.ErrorResp			"订单不存在"
//	@Router			/orders/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, ecode.ErrInvalidParams.WithDetails("id must be a positive integer"))
		return
	}
	o, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, toOrderResp(o))
}

// List godoc
//
//	@Summary		分页查询订单
//	@Tags			order
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int	false	"页码,从 1 开始"		example(1)
//	@Param			pageSize	query		int	false	"每页数量,最大 100"	example(20)
//	@Success		200			{object}	pkgdto.ListResp[dto.OrderResp]	"查询成功"
//	@Failure		400			{object}	pkgdto.ErrorResp				"参数无效"
//	@Failure		401			{object}	pkgdto.ErrorResp				"未认证"
//	@Router			/orders [get]
func (h *Handler) List(c *gin.Context) {
	var req dto.ListOrderReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, ecode.ErrInvalidParams.WithDetails(err.Error()))
		return
	}

	page, pageSize := pkgdto.NormalizePage(req.Page, req.PageSize)
	orders, total, err := h.svc.List(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}

	list := make([]dto.OrderResp, 0, len(orders))
	for _, o := range orders {
		list = append(list, toOrderResp(o))
	}
	response.Page(c, list, total, page, pageSize)
}

// Cancel godoc
//
//	@Summary		取消订单
//	@Description	仅待支付订单可以取消
//	@Tags			order
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		integer						true	"订单 ID"	example(1)
//	@Success		200	{object}	pkgdto.Resp[dto.OrderResp]	"取消成功"
//	@Failure		400	{object}	pkgdto.ErrorResp			"参数无效"
//	@Failure		401	{object}	pkgdto.ErrorResp			"未认证"
//	@Failure		404	{object}	pkgdto.ErrorResp			"订单不存在"
//	@Failure		409	{object}	pkgdto.ErrorResp			"订单状态不允许取消"
//	@Router			/orders/{id}/cancel [post]
func (h *Handler) Cancel(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, ecode.ErrInvalidParams.WithDetails("id must be a positive integer"))
		return
	}
	o, err := h.svc.Cancel(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, toOrderResp(o))
}

// parseID 解析 :id 路径参数为无符号整数。
func parseID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("id"), 10, 64)
}

func toOrderResp(o *model.Order) dto.OrderResp {
	return dto.OrderResp{
		ID:        o.ID,
		UserID:    o.UserID,
		OrderNo:   o.OrderNo,
		Amount:    o.Amount,
		Status:    o.Status,
		CreatedAt: o.CreatedAt.Format(time.RFC3339),
	}
}

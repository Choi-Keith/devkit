package orderhttp

import (
	"context"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/Choi-Keith/golang-common/ecode"
	"github.com/Choi-Keith/golang-common/response"

	"github.com/Choi-Keith/golang-ddd-template/internal/application/order"
	"github.com/Choi-Keith/golang-ddd-template/internal/domain/order"
)

// OrderUseCase is the slice of the application service the handler needs.
// Defined here so the handler stays testable with a fake implementation.
type OrderUseCase interface {
	Create(ctx context.Context, in orderapp.CreateInput) (*orderapp.OrderDTO, error)
	GetByID(ctx context.Context, id uint64) (*orderapp.OrderDTO, error)
	List(ctx context.Context, page, pageSize int) (*orderapp.OrderListDTO, error)
	Cancel(ctx context.Context, id uint64) (*orderapp.OrderDTO, error)
}

// OrderHandler exposes the order use cases over HTTP.
type OrderHandler struct {
	svc OrderUseCase
}

// NewOrderHandler creates an OrderHandler.
func NewOrderHandler(svc OrderUseCase) *OrderHandler {
	return &OrderHandler{svc: svc}
}

// Create godoc
//
//	@Summary		创建订单
//	@Description	为指定用户创建一个待支付订单,订单号由服务端生成
//	@Tags			order
//	@Accept			json
//	@Produce		json
//	@Param			request	body		CreateOrderRequest	true	"创建订单请求"
//	@Success		201		{object}	OrderEnvelope		"创建成功"
//	@Failure		400		{object}	ErrorEnvelope		"参数无效"
//	@Router			/orders [post]
func (h *OrderHandler) Create(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, ecode.ErrInvalidParams.WithDetails(err.Error()))
		return
	}

	out, err := h.svc.Create(c.Request.Context(), orderapp.CreateInput{
		UserID: req.UserID,
		Amount: req.Amount,
	})
	if err != nil {
		handleError(c, err)
		return
	}
	response.Created(c, toOrderResponse(out))
}

// GetByID godoc
//
//	@Summary		查询订单详情
//	@Tags			order
//	@Produce		json
//	@Param			id	path		integer			true	"订单 ID"	example(1)
//	@Success		200	{object}	OrderEnvelope	"查询成功"
//	@Failure		400	{object}	ErrorEnvelope	"参数无效"
//	@Failure		404	{object}	ErrorEnvelope	"订单不存在"
//	@Router			/orders/{id} [get]
func (h *OrderHandler) GetByID(c *gin.Context) {
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
	response.Success(c, toOrderResponse(out))
}

// List godoc
//
//	@Summary		分页查询订单
//	@Tags			order
//	@Produce		json
//	@Param			page		query		int	false	"页码,从 1 开始"		example(1)
//	@Param			pageSize	query		int	false	"每页数量,最大 100"	example(20)
//	@Success		200			{object}	OrderListEnvelope	"查询成功"
//	@Failure		400			{object}	ErrorEnvelope		"参数无效"
//	@Router			/orders [get]
func (h *OrderHandler) List(c *gin.Context) {
	var req ListOrderRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, ecode.ErrInvalidParams.WithDetails(err.Error()))
		return
	}

	out, err := h.svc.List(c.Request.Context(), req.Page, req.PageSize)
	if err != nil {
		handleError(c, err)
		return
	}

	list := make([]OrderResponse, 0, len(out.List))
	for i := range out.List {
		list = append(list, toOrderResponse(&out.List[i]))
	}
	response.Page(c, list, out.Total, out.Page, out.PageSize)
}

// Cancel godoc
//
//	@Summary		取消订单
//	@Description	仅待支付订单可以取消
//	@Tags			order
//	@Produce		json
//	@Param			id	path		integer			true	"订单 ID"	example(1)
//	@Success		200	{object}	OrderEnvelope	"取消成功"
//	@Failure		400	{object}	ErrorEnvelope	"参数无效"
//	@Failure		404	{object}	ErrorEnvelope	"订单不存在"
//	@Failure		409	{object}	ErrorEnvelope	"订单状态不允许取消"
//	@Router			/orders/{id}/cancel [post]
func (h *OrderHandler) Cancel(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, ecode.ErrInvalidParams.WithDetails("id must be a positive integer"))
		return
	}
	out, err := h.svc.Cancel(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, toOrderResponse(out))
}

// parseID parses the :id path parameter as a bigint unsigned id.
func parseID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("id"), 10, 64)
}

// handleError maps domain errors to business error codes at the boundary.
func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, order.ErrNotFound):
		response.Error(c, ErrOrderNotFound)
	case errors.Is(err, order.ErrNotCancellable):
		response.Error(c, ErrOrderNotCancellable)
	case errors.Is(err, order.ErrUserIDRequired),
		errors.Is(err, order.ErrOrderNoRequired),
		errors.Is(err, order.ErrInvalidAmount):
		response.Error(c, ecode.ErrInvalidParams.WithDetails(err.Error()))
	default:
		response.Error(c, ecode.ErrInternalServer.WithDetails(err.Error()))
	}
}

func toOrderResponse(d *orderapp.OrderDTO) OrderResponse {
	return OrderResponse{
		ID:        d.ID,
		UserID:    d.UserID,
		OrderNo:   d.OrderNo,
		Amount:    d.Amount,
		Status:    d.Status,
		CreatedAt: d.CreatedAt,
	}
}

package orderhttp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/Choi-Keith/golang-ddd-template/internal/application/order"
	"github.com/Choi-Keith/golang-ddd-template/internal/domain/order"
)

// fakeUseCase implements OrderUseCase for handler tests.
type fakeUseCase struct {
	createErr error
	getErr    error
	cancelErr error
}

func (f *fakeUseCase) Create(_ context.Context, in orderapp.CreateInput) (*orderapp.OrderDTO, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	return &orderapp.OrderDTO{ID: 1, UserID: in.UserID, OrderNo: "O123", Amount: in.Amount, Status: "pending", CreatedAt: "2026-08-27T08:00:00Z"}, nil
}

func (f *fakeUseCase) GetByID(_ context.Context, id uint64) (*orderapp.OrderDTO, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return &orderapp.OrderDTO{ID: id, UserID: 1, OrderNo: "O123", Amount: 199900, Status: "pending", CreatedAt: "2026-08-27T08:00:00Z"}, nil
}

func (f *fakeUseCase) List(_ context.Context, page, pageSize int) (*orderapp.OrderListDTO, error) {
	return &orderapp.OrderListDTO{List: []orderapp.OrderDTO{}, Total: 0, Page: page, PageSize: pageSize}, nil
}

func (f *fakeUseCase) Cancel(_ context.Context, id uint64) (*orderapp.OrderDTO, error) {
	if f.cancelErr != nil {
		return nil, f.cancelErr
	}
	return &orderapp.OrderDTO{ID: id, UserID: 1, OrderNo: "O123", Amount: 199900, Status: "cancelled", CreatedAt: "2026-08-27T08:00:00Z"}, nil
}

type envelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func newTestRouter(uc OrderUseCase) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewOrderHandler(uc)
	r.POST("/api/v1/orders", h.Create)
	r.GET("/api/v1/orders/:id", h.GetByID)
	r.POST("/api/v1/orders/:id/cancel", h.Cancel)
	return r
}

func TestOrderHandler_Create(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		createErr  error
		wantStatus int
		wantCode   int
	}{
		{name: "ok", body: `{"userId":1,"amount":199900}`, wantStatus: http.StatusCreated, wantCode: 0},
		{name: "missing amount", body: `{"userId":1}`, wantStatus: http.StatusBadRequest, wantCode: 10000},
		{name: "invalid amount", body: `{"userId":1,"amount":199900}`, createErr: order.ErrInvalidAmount, wantStatus: http.StatusBadRequest, wantCode: 10000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newTestRouter(&fakeUseCase{createErr: tt.createErr})
			req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("want status %d, got %d: %s", tt.wantStatus, rec.Code, rec.Body.String())
			}
			var env envelope
			if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
				t.Fatalf("invalid json: %v", err)
			}
			if env.Code != tt.wantCode {
				t.Fatalf("want code %d, got %d", tt.wantCode, env.Code)
			}
		})
	}
}

func TestOrderHandler_GetByID(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		getErr     error
		wantStatus int
		wantCode   int
	}{
		{name: "ok", path: "/api/v1/orders/1", wantStatus: http.StatusOK, wantCode: 0},
		{name: "invalid id", path: "/api/v1/orders/abc", wantStatus: http.StatusBadRequest, wantCode: 10000},
		{name: "not found", path: "/api/v1/orders/9999", getErr: order.ErrNotFound, wantStatus: http.StatusNotFound, wantCode: 20100},
		{name: "internal", path: "/api/v1/orders/1", getErr: errors.New("db down"), wantStatus: http.StatusInternalServerError, wantCode: 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newTestRouter(&fakeUseCase{getErr: tt.getErr})
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("want status %d, got %d: %s", tt.wantStatus, rec.Code, rec.Body.String())
			}
			var env envelope
			if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
				t.Fatalf("invalid json: %v", err)
			}
			if env.Code != tt.wantCode {
				t.Fatalf("want code %d, got %d", tt.wantCode, env.Code)
			}
		})
	}
}

func TestOrderHandler_Cancel(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		cancelErr  error
		wantStatus int
		wantCode   int
	}{
		{name: "ok", path: "/api/v1/orders/1/cancel", wantStatus: http.StatusOK, wantCode: 0},
		{name: "invalid id", path: "/api/v1/orders/abc/cancel", wantStatus: http.StatusBadRequest, wantCode: 10000},
		{name: "not found", path: "/api/v1/orders/9999/cancel", cancelErr: order.ErrNotFound, wantStatus: http.StatusNotFound, wantCode: 20100},
		{name: "not cancellable", path: "/api/v1/orders/1/cancel", cancelErr: order.ErrNotCancellable, wantStatus: http.StatusConflict, wantCode: 20101},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newTestRouter(&fakeUseCase{cancelErr: tt.cancelErr})
			req := httptest.NewRequest(http.MethodPost, tt.path, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("want status %d, got %d: %s", tt.wantStatus, rec.Code, rec.Body.String())
			}
			var env envelope
			if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
				t.Fatalf("invalid json: %v", err)
			}
			if env.Code != tt.wantCode {
				t.Fatalf("want code %d, got %d", tt.wantCode, env.Code)
			}
		})
	}
}

package userhttp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/Choi-Keith/golang-ddd-template/internal/application/user"
	"github.com/Choi-Keith/golang-ddd-template/internal/domain/user"
)

// fakeUseCase implements UserUseCase for handler tests.
type fakeUseCase struct {
	registerErr error
	getErr      error
}

func (f *fakeUseCase) Register(_ context.Context, in userapp.RegisterInput) (*userapp.UserDTO, error) {
	if f.registerErr != nil {
		return nil, f.registerErr
	}
	return &userapp.UserDTO{ID: 1, Email: in.Email, Name: in.Name, CreatedAt: "2026-08-27T08:00:00Z"}, nil
}

func (f *fakeUseCase) GetByID(_ context.Context, id uint64) (*userapp.UserDTO, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return &userapp.UserDTO{ID: id, Email: "a@b.com", Name: "Alice", CreatedAt: "2026-08-27T08:00:00Z"}, nil
}

func (f *fakeUseCase) List(_ context.Context, page, pageSize int) (*userapp.UserListDTO, error) {
	return &userapp.UserListDTO{List: []userapp.UserDTO{}, Total: 0, Page: page, PageSize: pageSize}, nil
}

func (f *fakeUseCase) Delete(_ context.Context, _ uint64) error { return nil }

type envelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func newTestRouter(uc UserUseCase) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewUserHandler(uc)
	r.POST("/api/v1/users", h.Register)
	r.GET("/api/v1/users/:id", h.GetByID)
	return r
}

func TestUserHandler_Register(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		registerErr error
		wantStatus  int
		wantCode    int
	}{
		{name: "ok", body: `{"email":"a@b.com","name":"Alice","password":"secret123"}`, wantStatus: http.StatusCreated, wantCode: 0},
		{name: "missing password", body: `{"email":"a@b.com","name":"Alice"}`, wantStatus: http.StatusBadRequest, wantCode: 10000},
		{name: "email taken", body: `{"email":"a@b.com","name":"Alice","password":"secret123"}`, registerErr: user.ErrEmailTaken, wantStatus: http.StatusConflict, wantCode: 20001},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newTestRouter(&fakeUseCase{registerErr: tt.registerErr})
			req := httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(tt.body))
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

func TestUserHandler_GetByID(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		getErr     error
		wantStatus int
		wantCode   int
	}{
		{name: "ok", path: "/api/v1/users/1", wantStatus: http.StatusOK, wantCode: 0},
		{name: "invalid id", path: "/api/v1/users/abc", wantStatus: http.StatusBadRequest, wantCode: 10000},
		{name: "not found", path: "/api/v1/users/9999", getErr: user.ErrNotFound, wantStatus: http.StatusNotFound, wantCode: 20000},
		{name: "internal", path: "/api/v1/users/1", getErr: errors.New("db down"), wantStatus: http.StatusInternalServerError, wantCode: 1000},
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

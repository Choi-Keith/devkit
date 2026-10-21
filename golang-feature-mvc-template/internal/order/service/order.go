// Package service 订单模块应用层:创建、查询、分页、取消。
// Repository 接口定义在本包(消费方),实现位于 internal/infra,
// 由 main 手工装配注入。业务失败直接返回 pkg/xerr 中的业务错误码
// (*ecode.Error),handler 原样交给 response.Error 即可。
package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"

	"gorm.io/gorm"

	"github.com/Choi-Keith/golang-common/ecode"

	"github.com/Choi-Keith/golang-feature-mvc-template/internal/infra/db/order/model"
	userservice "github.com/Choi-Keith/golang-feature-mvc-template/internal/user/service"
	pkgdto "github.com/Choi-Keith/golang-feature-mvc-template/pkg/dto"
	"github.com/Choi-Keith/golang-feature-mvc-template/pkg/xerr"
)

// Repository 订单数据访问端口。查询/更新类方法在记录不存在时必须返回
// gorm.ErrRecordNotFound,由本包翻译为 pkg/xerr 中的业务错误码。
type Repository interface {
	Create(ctx context.Context, o *model.Order) error
	FindByID(ctx context.Context, id uint64) (*model.Order, error)
	List(ctx context.Context, page, pageSize int) ([]*model.Order, int64, error)
	Update(ctx context.Context, o *model.Order) error
}

// Service 订单业务逻辑。对 user 模块的依赖是单向的:只允许调用
// user service 暴露的方法做用户存在性校验,禁止 import user 的
// repository / model。
type Service struct {
	repo  Repository
	users *userservice.Service
}

// NewService 创建订单服务。
func NewService(repo Repository, users *userservice.Service) *Service {
	return &Service{repo: repo, users: users}
}

// Create 创建待支付订单,订单号由服务端生成(时间戳保序 + 随机位防碰撞)。
func (s *Service) Create(ctx context.Context, userID uint64, amount int64) (*model.Order, error) {
	if userID == 0 {
		return nil, ecode.ErrInvalidParams.WithDetails("user id required")
	}
	if amount <= 0 {
		return nil, ecode.ErrInvalidParams.WithDetails("amount must be positive")
	}

	exists, err := s.users.Exists(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("order.service.Create: check user: %w", err)
	}
	if !exists {
		return nil, xerr.ErrOrderUserNotFound
	}

	o := &model.Order{
		UserID:  userID,
		OrderNo: fmt.Sprintf("O%d%04d", time.Now().UnixNano(), rand.IntN(10000)),
		Amount:  amount,
		Status:  model.StatusPending,
	}
	if err := s.repo.Create(ctx, o); err != nil {
		return nil, fmt.Errorf("order.service.Create: %w", err)
	}
	return o, nil
}

// GetByID 查询单个订单。
func (s *Service) GetByID(ctx context.Context, id uint64) (*model.Order, error) {
	o, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("order.service.GetByID: %w", mapRepoErr(err))
	}
	return o, nil
}

// List 分页查询订单。
func (s *Service) List(ctx context.Context, page, pageSize int) ([]*model.Order, int64, error) {
	page, pageSize = pkgdto.NormalizePage(page, pageSize)
	orders, total, err := s.repo.List(ctx, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("order.service.List: %w", err)
	}
	return orders, total, nil
}

// Cancel 取消待支付订单,非 pending 状态返回 xerr.ErrOrderNotCancellable。
func (s *Service) Cancel(ctx context.Context, id uint64) (*model.Order, error) {
	o, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("order.service.Cancel: %w", mapRepoErr(err))
	}
	if o.Status != model.StatusPending {
		return nil, xerr.ErrOrderNotCancellable
	}
	o.Status = model.StatusCancelled
	if err := s.repo.Update(ctx, o); err != nil {
		return nil, fmt.Errorf("order.service.Cancel: %w", mapRepoErr(err))
	}
	return o, nil
}

// mapRepoErr 把仓储层的 GORM 错误词汇翻译为 pkg/xerr 业务错误码。
func mapRepoErr(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return xerr.ErrOrderNotFound
	}
	return err
}

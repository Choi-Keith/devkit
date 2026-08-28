// Package service 业务逻辑层,接口 + 实现。
package service

import (
	"context"

	"github.com/google/uuid"

	"golang-simple-template/internal/ecode"
	"golang-simple-template/internal/model"
	"golang-simple-template/internal/repository"
	"golang-simple-template/internal/types"
)

// UserService 用户业务接口
type UserService interface {
	Create(ctx context.Context, req *types.CreateUserRequest) (*types.UserResponse, error)
	GetByID(ctx context.Context, id string) (*types.UserResponse, error)
	List(ctx context.Context, page, pageSize int) ([]types.UserResponse, int64, error)
	Update(ctx context.Context, id string, req *types.UpdateUserRequest) (*types.UserResponse, error)
	Delete(ctx context.Context, id string) error
}

type userService struct {
	userRepo repository.UserRepository
}

// NewUserService 创建用户服务
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) Create(ctx context.Context, req *types.CreateUserRequest) (*types.UserResponse, error) {
	if _, err := s.userRepo.GetByEmail(ctx, req.Email); err == nil {
		return nil, ecode.ErrEmailExists
	} else if !repository.IsNotFound(err) {
		return nil, err
	}

	user := &model.User{
		ID:    uuid.NewString(),
		Email: req.Email,
		Name:  req.Name,
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}
	return toUserResponse(user), nil
}

func (s *userService) GetByID(ctx context.Context, id string) (*types.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if repository.IsNotFound(err) {
			return nil, ecode.ErrUserNotFound
		}
		return nil, err
	}
	return toUserResponse(user), nil
}

func (s *userService) List(ctx context.Context, page, pageSize int) ([]types.UserResponse, int64, error) {
	users, total, err := s.userRepo.List(ctx, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	resp := make([]types.UserResponse, 0, len(users))
	for i := range users {
		resp = append(resp, *toUserResponse(&users[i]))
	}
	return resp, total, nil
}

func (s *userService) Update(ctx context.Context, id string, req *types.UpdateUserRequest) (*types.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if repository.IsNotFound(err) {
			return nil, ecode.ErrUserNotFound
		}
		return nil, err
	}
	user.Name = req.Name
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}
	return toUserResponse(user), nil
}

func (s *userService) Delete(ctx context.Context, id string) error {
	if _, err := s.GetByID(ctx, id); err != nil {
		return err
	}
	return s.userRepo.Delete(ctx, id)
}

func toUserResponse(u *model.User) *types.UserResponse {
	return &types.UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

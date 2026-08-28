package userapp

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/Choi-Keith/golang-ddd-template/internal/domain/user"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

// UserService orchestrates the user use cases.
type UserService struct {
	repo user.Repository
}

// NewUserService creates a UserService backed by the given repository.
func NewUserService(repo user.Repository) *UserService {
	return &UserService{repo: repo}
}

// Register creates a new user account.
func (s *UserService) Register(ctx context.Context, in RegisterInput) (*UserDTO, error) {
	existing, err := s.repo.FindByEmail(ctx, in.Email)
	if err != nil && !errors.Is(err, user.ErrNotFound) {
		return nil, fmt.Errorf("userapp.Register: find by email: %w", err)
	}
	if existing != nil {
		return nil, user.ErrEmailTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("userapp.Register: hash password: %w", err)
	}

	u, err := user.NewUser(in.Email, in.Name, string(hash))
	if err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, u); err != nil {
		return nil, fmt.Errorf("userapp.Register: save user: %w", err)
	}
	return toDTO(u), nil
}

// GetByID returns a single user.
func (s *UserService) GetByID(ctx context.Context, id uint64) (*UserDTO, error) {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("userapp.GetByID: %w", err)
	}
	return toDTO(u), nil
}

// List returns a page of users.
func (s *UserService) List(ctx context.Context, page, pageSize int) (*UserListDTO, error) {
	page, pageSize = normalizePage(page, pageSize)

	users, total, err := s.repo.List(ctx, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("userapp.List: %w", err)
	}

	list := make([]UserDTO, 0, len(users))
	for _, u := range users {
		list = append(list, *toDTO(u))
	}
	return &UserListDTO{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

// Delete removes a user.
func (s *UserService) Delete(ctx context.Context, id uint64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("userapp.Delete: %w", err)
	}
	return nil
}

func normalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return page, pageSize
}

func toDTO(u *user.User) *UserDTO {
	return &UserDTO{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
	}
}

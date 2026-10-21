// Package service 用户模块应用层:注册、登录、令牌刷新/吊销与用户 CRUD。
// Repository / Cache 接口定义在本包(消费方),实现位于 internal/infra,
// 由 main 手工装配注入。业务失败直接返回 pkg/xerr 中的业务错误码
// (*ecode.Error),handler 原样交给 response.Error 即可。
package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/Choi-Keith/golang-common/auth"
	"github.com/Choi-Keith/golang-common/logger"

	"github.com/Choi-Keith/golang-feature-mvc-template/internal/infra/db/user/model"
	pkgdto "github.com/Choi-Keith/golang-feature-mvc-template/pkg/dto"
	"github.com/Choi-Keith/golang-feature-mvc-template/pkg/xerr"
)

// Repository 用户数据访问端口。查询类方法在记录不存在时必须返回
// gorm.ErrRecordNotFound,唯一键冲突(邮箱重复)时必须返回
// gorm.ErrDuplicatedKey,均由本包翻译为 pkg/xerr 中的业务错误码。
type Repository interface {
	Create(ctx context.Context, u *model.User) error
	FindByID(ctx context.Context, id uint64) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	List(ctx context.Context, page, pageSize int) ([]*model.User, int64, error)
	Update(ctx context.Context, u *model.User) error
	Delete(ctx context.Context, id uint64) error
}

// Cache 缓存端口:用户详情 cache-aside 与 refresh token 吊销列表。
// 传 nil 时降级为直连 DB、不做 refresh token 吊销校验。
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
}

// TokenManager 抽象 JWT 签发与校验,*auth.JWT 即实现,测试可替换。
type TokenManager interface {
	GenerateToken(userID, email string) (string, error)
	ValidateToken(tokenString string) (*auth.JWTClaims, error)
}

// Service 用户业务逻辑。
type Service struct {
	repo         Repository
	cache        Cache
	accessToken  TokenManager
	refreshToken TokenManager
	refreshTTL   time.Duration
	cacheTTL     time.Duration
}

// NewService 创建用户服务。accessToken / refreshToken 只是过期时间不同的
// 两个 JWT 管理器;refreshTTL 必须与 refreshToken 的过期时间一致。
func NewService(repo Repository, cache Cache, accessToken, refreshToken TokenManager, refreshTTL time.Duration) *Service {
	return &Service{
		repo:         repo,
		cache:        cache,
		accessToken:  accessToken,
		refreshToken: refreshToken,
		refreshTTL:   refreshTTL,
		cacheTTL:     10 * time.Minute,
	}
}

// Register 创建新用户,邮箱已注册时返回 xerr.ErrEmailTaken。
func (s *Service) Register(ctx context.Context, email, name, password string) (*model.User, error) {
	_, err := s.repo.FindByEmail(ctx, email)
	switch {
	case err == nil:
		return nil, xerr.ErrEmailTaken
	case !errors.Is(err, gorm.ErrRecordNotFound):
		return nil, fmt.Errorf("user.service.Register: find by email: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("user.service.Register: hash password: %w", err)
	}

	u := &model.User{Email: email, Name: name, Password: string(hash)}
	if err := s.repo.Create(ctx, u); err != nil {
		// 并发下的唯一键冲突(先查后写的竞态)视为邮箱已注册。
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, xerr.ErrEmailTaken
		}
		return nil, fmt.Errorf("user.service.Register: %w", err)
	}
	return u, nil
}

// Login 校验邮箱密码,成功后签发 access/refresh token,并把 refresh token
// 的哈希写入 Redis(logout 吊销与 refresh 轮换的依据)。
// 为避免用户枚举,邮箱不存在与密码错误返回同一个 xerr.ErrInvalidCredential。
func (s *Service) Login(ctx context.Context, email, password string) (access, refresh string, err error) {
	u, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", xerr.ErrInvalidCredential
		}
		return "", "", fmt.Errorf("user.service.Login: find by email: %w", err)
	}
	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) != nil {
		return "", "", xerr.ErrInvalidCredential
	}
	return s.issueTokenPair(ctx, u)
}

// Refresh 用 refresh token 换新的一对 token。旧 token 的哈希必须与 Redis
// 中存储的一致,否则视为已吊销;成功后轮换哈希,旧 token 立即失效。
func (s *Service) Refresh(ctx context.Context, refreshToken string) (access, refresh string, err error) {
	claims, err := s.refreshToken.ValidateToken(refreshToken)
	if err != nil {
		return "", "", xerr.ErrRefreshTokenInvalid
	}
	userID, err := strconv.ParseUint(claims.UserID, 10, 64)
	if err != nil {
		return "", "", xerr.ErrRefreshTokenInvalid
	}

	if s.cache != nil {
		stored, err := s.cache.Get(ctx, refreshKey(userID))
		if err != nil || stored != hashToken(refreshToken) {
			return "", "", xerr.ErrRefreshTokenRevoked
		}
	}

	u, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return "", "", fmt.Errorf("user.service.Refresh: %w", mapRepoErr(err))
	}
	return s.issueTokenPair(ctx, u)
}

// Logout 吊销 refresh token:删除 Redis 中的哈希。token 本身非法时返回
// xerr.ErrRefreshTokenInvalid;已吊销视为成功(幂等)。
func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	claims, err := s.refreshToken.ValidateToken(refreshToken)
	if err != nil {
		return xerr.ErrRefreshTokenInvalid
	}
	userID, err := strconv.ParseUint(claims.UserID, 10, 64)
	if err != nil {
		return xerr.ErrRefreshTokenInvalid
	}
	if s.cache != nil {
		if err := s.cache.Del(ctx, refreshKey(userID)); err != nil {
			return fmt.Errorf("user.service.Logout: revoke token: %w", err)
		}
	}
	return nil
}

// GetByID 查询用户详情,演示 cache-aside:先查 Redis,miss 走 DB 并回写;
// 缓存故障降级为直连 DB,仅记录日志。
func (s *Service) GetByID(ctx context.Context, id uint64) (*model.User, error) {
	key := detailKey(id)
	if s.cache != nil {
		if data, err := s.cache.Get(ctx, key); err == nil {
			var u model.User
			if json.Unmarshal([]byte(data), &u) == nil {
				return &u, nil
			}
		}
	}

	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user.service.GetByID: %w", mapRepoErr(err))
	}

	if s.cache != nil {
		if data, err := json.Marshal(u); err == nil {
			if err := s.cache.Set(ctx, key, string(data), s.cacheTTL); err != nil {
				logger.Warn("user.service.GetByID: write cache", logger.UserID(strconv.FormatUint(id, 10)))
			}
		}
	}
	return u, nil
}

// List 分页查询用户。
func (s *Service) List(ctx context.Context, page, pageSize int) ([]*model.User, int64, error) {
	page, pageSize = pkgdto.NormalizePage(page, pageSize)
	users, total, err := s.repo.List(ctx, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("user.service.List: %w", err)
	}
	return users, total, nil
}

// Update 修改昵称,password 非空时一并重置密码。成功后使详情缓存失效。
func (s *Service) Update(ctx context.Context, id uint64, name, password string) (*model.User, error) {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user.service.Update: %w", mapRepoErr(err))
	}

	if name != "" {
		u.Name = name
	}
	if password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("user.service.Update: hash password: %w", err)
		}
		u.Password = string(hash)
	}

	if err := s.repo.Update(ctx, u); err != nil {
		return nil, fmt.Errorf("user.service.Update: %w", mapRepoErr(err))
	}
	s.invalidateCache(ctx, id)
	return u, nil
}

// Delete 删除用户,同时使详情缓存与登录态失效。
func (s *Service) Delete(ctx context.Context, id uint64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("user.service.Delete: %w", mapRepoErr(err))
	}
	s.invalidateCache(ctx, id)
	return nil
}

// Exists 报告用户是否存在,供 order 等其他模块做存在性校验。
// 跨模块只允许调用本包暴露的方法,严禁 import 对方的 repository / model。
func (s *Service) Exists(ctx context.Context, id uint64) (bool, error) {
	_, err := s.repo.FindByID(ctx, id)
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		return false, nil
	default:
		return false, fmt.Errorf("user.service.Exists: %w", err)
	}
}

// issueTokenPair 签发一对 token 并存储 refresh token 哈希。
func (s *Service) issueTokenPair(ctx context.Context, u *model.User) (access, refresh string, err error) {
	userID := strconv.FormatUint(u.ID, 10)

	access, err = s.accessToken.GenerateToken(userID, u.Email)
	if err != nil {
		return "", "", fmt.Errorf("user.service: generate access token: %w", err)
	}
	refresh, err = s.refreshToken.GenerateToken(userID, u.Email)
	if err != nil {
		return "", "", fmt.Errorf("user.service: generate refresh token: %w", err)
	}

	if s.cache != nil {
		if err := s.cache.Set(ctx, refreshKey(u.ID), hashToken(refresh), s.refreshTTL); err != nil {
			return "", "", fmt.Errorf("user.service: store refresh token: %w", err)
		}
	}
	return access, refresh, nil
}

func (s *Service) invalidateCache(ctx context.Context, id uint64) {
	if s.cache == nil {
		return
	}
	if err := s.cache.Del(ctx, detailKey(id), refreshKey(id)); err != nil {
		logger.Warn("user.service: invalidate cache", logger.UserID(strconv.FormatUint(id, 10)))
	}
}

// detailKey 用户详情缓存 key。
func detailKey(id uint64) string {
	return fmt.Sprintf("user:detail:%d", id)
}

// refreshKey 用户 refresh token 哈希 key,一个用户同时只有一条有效登录态。
func refreshKey(id uint64) string {
	return fmt.Sprintf("user:refresh:%d", id)
}

// hashToken 只存 token 的 SHA-256 哈希,Redis 泄露也无法伪造登录态。
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// mapRepoErr 把仓储层的 GORM 错误词汇翻译为 pkg/xerr 业务错误码。
func mapRepoErr(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return xerr.ErrUserNotFound
	}
	return err
}

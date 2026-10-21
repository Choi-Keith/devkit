package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/Choi-Keith/golang-common/auth"

	"github.com/Choi-Keith/golang-feature-mvc-template/internal/infra/db/user/model"
	"github.com/Choi-Keith/golang-feature-mvc-template/pkg/xerr"
)

// fakeRepo 是 Repository 接口的内存测试替身。
type fakeRepo struct {
	nextID  uint64
	byID    map[uint64]*model.User
	byEmail map[string]*model.User
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{nextID: 1, byID: map[uint64]*model.User{}, byEmail: map[string]*model.User{}}
}

func (f *fakeRepo) Create(_ context.Context, u *model.User) error {
	if _, ok := f.byEmail[u.Email]; ok {
		return gorm.ErrDuplicatedKey
	}
	u.ID = f.nextID // 模拟自增主键
	u.CreatedAt = time.Now()
	u.UpdatedAt = u.CreatedAt
	f.nextID++
	f.byID[u.ID] = u
	f.byEmail[u.Email] = u
	return nil
}

func (f *fakeRepo) FindByID(_ context.Context, id uint64) (*model.User, error) {
	if u, ok := f.byID[id]; ok {
		return u, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeRepo) FindByEmail(_ context.Context, email string) (*model.User, error) {
	if u, ok := f.byEmail[email]; ok {
		return u, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeRepo) List(_ context.Context, page, pageSize int) ([]*model.User, int64, error) {
	all := make([]*model.User, 0, len(f.byID))
	for _, u := range f.byID {
		all = append(all, u)
	}
	start := (page - 1) * pageSize
	if start >= len(all) {
		return nil, int64(len(all)), nil
	}
	end := start + pageSize
	if end > len(all) {
		end = len(all)
	}
	return all[start:end], int64(len(all)), nil
}

func (f *fakeRepo) Update(_ context.Context, u *model.User) error {
	if _, ok := f.byID[u.ID]; !ok {
		return gorm.ErrRecordNotFound
	}
	f.byID[u.ID] = u
	f.byEmail[u.Email] = u
	return nil
}

func (f *fakeRepo) Delete(_ context.Context, id uint64) error {
	u, ok := f.byID[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	delete(f.byID, id)
	delete(f.byEmail, u.Email)
	return nil
}

// fakeCache 是 Cache 接口的内存测试替身。
type fakeCache struct {
	m map[string]string
}

func newFakeCache() *fakeCache {
	return &fakeCache{m: map[string]string{}}
}

func (f *fakeCache) Get(_ context.Context, key string) (string, error) {
	if v, ok := f.m[key]; ok {
		return v, nil
	}
	return "", fmt.Errorf("cache miss: %s", key)
}

func (f *fakeCache) Set(_ context.Context, key string, value string, _ time.Duration) error {
	f.m[key] = value
	return nil
}

func (f *fakeCache) Del(_ context.Context, keys ...string) error {
	for _, k := range keys {
		delete(f.m, k)
	}
	return nil
}

// newTestService 用真实 JWT 管理器 + 内存 repo/cache 组装被测服务。
func newTestService() (*Service, *fakeRepo, *fakeCache) {
	repo := newFakeRepo()
	c := newFakeCache()
	access := auth.NewJWT(auth.JWTConfig{Secret: "test-secret", Expiration: 2 * time.Hour})
	refresh := auth.NewJWT(auth.JWTConfig{Secret: "test-secret", Expiration: 7 * 24 * time.Hour})
	return NewService(repo, c, access, refresh, 7*24*time.Hour), repo, c
}

func TestService_Register(t *testing.T) {
	svc, _, _ := newTestService()

	u, err := svc.Register(context.Background(), "alice@example.com", "Alice", "P@ssw0rd123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.ID == 0 {
		t.Fatal("id must be assigned after create")
	}
	if u.Password == "P@ssw0rd123" || u.Password == "" {
		t.Fatal("password must be bcrypt-hashed")
	}

	// 邮箱重复注册必须失败(含先查后写竞态下的唯一键冲突)。
	if _, err := svc.Register(context.Background(), "alice@example.com", "Alice2", "P@ssw0rd123"); !errors.Is(err, xerr.ErrEmailTaken) {
		t.Fatalf("want xerr.ErrEmailTaken, got %v", err)
	}
}

func TestService_Login(t *testing.T) {
	svc, _, _ := newTestService()
	if _, err := svc.Register(context.Background(), "alice@example.com", "Alice", "P@ssw0rd123"); err != nil {
		t.Fatalf("setup: %v", err)
	}

	access, refresh, err := svc.Login(context.Background(), "alice@example.com", "P@ssw0rd123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if access == "" || refresh == "" {
		t.Fatal("login must return both tokens")
	}

	// 密码错误与邮箱不存在返回同一错误,避免用户枚举。
	if _, _, err := svc.Login(context.Background(), "alice@example.com", "wrong-password"); !errors.Is(err, xerr.ErrInvalidCredential) {
		t.Fatalf("want xerr.ErrInvalidCredential, got %v", err)
	}
	if _, _, err := svc.Login(context.Background(), "nobody@example.com", "P@ssw0rd123"); !errors.Is(err, xerr.ErrInvalidCredential) {
		t.Fatalf("want xerr.ErrInvalidCredential, got %v", err)
	}
}

func TestService_RefreshAndLogout(t *testing.T) {
	svc, _, _ := newTestService()
	if _, err := svc.Register(context.Background(), "alice@example.com", "Alice", "P@ssw0rd123"); err != nil {
		t.Fatalf("setup: %v", err)
	}
	_, refresh, err := svc.Login(context.Background(), "alice@example.com", "P@ssw0rd123")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	// 正常刷新会轮换令牌。
	access2, refresh2, err := svc.Refresh(context.Background(), refresh)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if access2 == "" || refresh2 == "" {
		t.Fatal("refresh must return a new token pair")
	}

	// 旧 refresh token 已轮换,再用必须视为吊销。
	if _, _, err := svc.Refresh(context.Background(), refresh); !errors.Is(err, xerr.ErrRefreshTokenRevoked) {
		t.Fatalf("want xerr.ErrRefreshTokenRevoked, got %v", err)
	}

	// 登出后新 token 也被吊销。
	if err := svc.Logout(context.Background(), refresh2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, _, err := svc.Refresh(context.Background(), refresh2); !errors.Is(err, xerr.ErrRefreshTokenRevoked) {
		t.Fatalf("want xerr.ErrRefreshTokenRevoked after logout, got %v", err)
	}

	// 登出幂等。
	if err := svc.Logout(context.Background(), refresh2); err != nil {
		t.Fatalf("logout must be idempotent: %v", err)
	}

	// 伪造 token 直接拒绝。
	if err := svc.Logout(context.Background(), "not-a-token"); !errors.Is(err, xerr.ErrRefreshTokenInvalid) {
		t.Fatalf("want xerr.ErrRefreshTokenInvalid, got %v", err)
	}
}

func TestService_GetByID(t *testing.T) {
	svc, _, c := newTestService()
	u, err := svc.Register(context.Background(), "alice@example.com", "Alice", "P@ssw0rd123")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	if _, err := svc.GetByID(context.Background(), 9999); !errors.Is(err, xerr.ErrUserNotFound) {
		t.Fatalf("want xerr.ErrUserNotFound, got %v", err)
	}

	got, err := svc.GetByID(context.Background(), u.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Email != u.Email {
		t.Fatalf("want email %s, got %s", u.Email, got.Email)
	}

	// cache-aside:第一次查询后应回写缓存。
	if _, err := c.Get(context.Background(), detailKey(u.ID)); err != nil {
		t.Fatalf("detail must be cached after GetByID: %v", err)
	}
}

func TestService_Update(t *testing.T) {
	svc, _, _ := newTestService()
	u, err := svc.Register(context.Background(), "alice@example.com", "Alice", "P@ssw0rd123")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	updated, err := svc.Update(context.Background(), u.ID, "Alice2", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != "Alice2" {
		t.Fatalf("want name Alice2, got %s", updated.Name)
	}

	// 旧密码仍应可用(未重置)。
	if _, _, err := svc.Login(context.Background(), "alice@example.com", "P@ssw0rd123"); err != nil {
		t.Fatalf("old password must still work: %v", err)
	}

	if _, err := svc.Update(context.Background(), u.ID, "", "N3wP@ssw0rd"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, _, err := svc.Login(context.Background(), "alice@example.com", "N3wP@ssw0rd"); err != nil {
		t.Fatalf("new password must work: %v", err)
	}

	if _, err := svc.Update(context.Background(), 9999, "x", ""); !errors.Is(err, xerr.ErrUserNotFound) {
		t.Fatalf("want xerr.ErrUserNotFound, got %v", err)
	}
}

func TestService_List(t *testing.T) {
	svc, _, _ := newTestService()
	for i := 0; i < 3; i++ {
		if _, err := svc.Register(context.Background(), fmt.Sprintf("user%d@example.com", i), "U", "P@ssw0rd123"); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}

	users, total, err := svc.List(context.Background(), 1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 3 || len(users) != 2 {
		t.Fatalf("want total=3 len=2, got total=%d len=%d", total, len(users))
	}
}

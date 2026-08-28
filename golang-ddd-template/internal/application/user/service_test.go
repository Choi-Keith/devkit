package userapp

import (
	"context"
	"errors"
	"testing"

	"github.com/Choi-Keith/golang-ddd-template/internal/domain/user"
)

// fakeRepo is an in-memory test double for user.Repository.
type fakeRepo struct {
	nextID  uint64
	byEmail map[string]*user.User
	byID    map[uint64]*user.User
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{nextID: 1, byEmail: map[string]*user.User{}, byID: map[uint64]*user.User{}}
}

func (f *fakeRepo) Save(_ context.Context, u *user.User) error {
	if _, ok := f.byEmail[u.Email]; ok {
		return user.ErrEmailTaken
	}
	u.ID = f.nextID // simulate auto-increment
	f.nextID++
	f.byEmail[u.Email] = u
	f.byID[u.ID] = u
	return nil
}

func (f *fakeRepo) FindByID(_ context.Context, id uint64) (*user.User, error) {
	if u, ok := f.byID[id]; ok {
		return u, nil
	}
	return nil, user.ErrNotFound
}

func (f *fakeRepo) FindByEmail(_ context.Context, email string) (*user.User, error) {
	if u, ok := f.byEmail[email]; ok {
		return u, nil
	}
	return nil, user.ErrNotFound
}

func (f *fakeRepo) List(_ context.Context, page, pageSize int) ([]*user.User, int64, error) {
	all := make([]*user.User, 0, len(f.byID))
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

func (f *fakeRepo) Delete(_ context.Context, id uint64) error {
	u, ok := f.byID[id]
	if !ok {
		return user.ErrNotFound
	}
	delete(f.byID, id)
	delete(f.byEmail, u.Email)
	return nil
}

func TestUserService_Register(t *testing.T) {
	tests := []struct {
		name    string
		in      RegisterInput
		seed    *user.User // pre-existing user, if any
		wantErr error
	}{
		{
			name: "ok",
			in:   RegisterInput{Email: "a@b.com", Name: "Alice", Password: "secret123"},
		},
		{
			name:    "duplicate email",
			in:      RegisterInput{Email: "a@b.com", Name: "Alice", Password: "secret123"},
			seed:    &user.User{ID: 9, Email: "a@b.com", Name: "Other", Password: "hash"},
			wantErr: user.ErrEmailTaken,
		},
		{
			name:    "invalid email",
			in:      RegisterInput{Email: "bad", Name: "Alice", Password: "secret123"},
			wantErr: user.ErrInvalidEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepo()
			if tt.seed != nil {
				repo.byEmail[tt.seed.Email] = tt.seed
				repo.byID[tt.seed.ID] = tt.seed
			}
			svc := NewUserService(repo)

			out, err := svc.Register(context.Background(), tt.in)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("want error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if out.ID == 0 {
				t.Fatal("id must be assigned after save")
			}
			if out.Email != tt.in.Email || out.Name != tt.in.Name {
				t.Fatalf("unexpected dto: %+v", out)
			}
			// Password must be hashed, never stored as plaintext.
			saved := repo.byID[out.ID]
			if saved.Password == tt.in.Password || saved.Password == "" {
				t.Fatal("password must be hashed before persistence")
			}
		})
	}
}

func TestUserService_GetByID(t *testing.T) {
	repo := newFakeRepo()
	svc := NewUserService(repo)

	created, err := svc.Register(context.Background(), RegisterInput{
		Email: "a@b.com", Name: "Alice", Password: "secret123",
	})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	if _, err := svc.GetByID(context.Background(), 9999); !errors.Is(err, user.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}

	got, err := svc.GetByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Email != created.Email {
		t.Fatalf("want email %s, got %s", created.Email, got.Email)
	}
}

func TestUserService_List(t *testing.T) {
	repo := newFakeRepo()
	svc := NewUserService(repo)

	for _, tc := range []struct{ email, name string }{
		{"a@b.com", "A"}, {"b@b.com", "B"}, {"c@b.com", "C"},
	} {
		if _, err := svc.Register(context.Background(), RegisterInput{
			Email: tc.email, Name: tc.name, Password: "secret123",
		}); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}

	out, err := svc.List(context.Background(), 1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Total != 3 || len(out.List) != 2 {
		t.Fatalf("want total=3 len=2, got total=%d len=%d", out.Total, len(out.List))
	}

	// Page normalization: invalid values fall back to defaults.
	out, err = svc.List(context.Background(), 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Page != 1 || out.PageSize != defaultPageSize {
		t.Fatalf("want page=1 pageSize=%d, got page=%d pageSize=%d", defaultPageSize, out.Page, out.PageSize)
	}
}

func TestUserService_Delete(t *testing.T) {
	repo := newFakeRepo()
	svc := NewUserService(repo)

	created, err := svc.Register(context.Background(), RegisterInput{
		Email: "a@b.com", Name: "Alice", Password: "secret123",
	})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	if err := svc.Delete(context.Background(), 9999); !errors.Is(err, user.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
	if err := svc.Delete(context.Background(), created.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

package user

import (
	"errors"
	"testing"
)

func TestNewUser(t *testing.T) {
	tests := []struct {
		name         string
		email        string
		userName     string
		passwordHash string
		wantErr      error
	}{
		{name: "ok", email: "a@b.com", userName: "Alice", passwordHash: "hash"},
		{name: "email required", email: "  ", userName: "Alice", passwordHash: "hash", wantErr: ErrEmailRequired},
		{name: "email invalid", email: "not-an-email", userName: "Alice", passwordHash: "hash", wantErr: ErrInvalidEmail},
		{name: "name required", email: "a@b.com", userName: "", passwordHash: "hash", wantErr: ErrNameRequired},
		{name: "password required", email: "a@b.com", userName: "Alice", passwordHash: "", wantErr: ErrPasswordRequired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := NewUser(tt.email, tt.userName, tt.passwordHash)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("want error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if u.ID != 0 {
				t.Fatalf("id must be zero before persistence, got %d", u.ID)
			}
			if u.Email != tt.email || u.Name != tt.userName {
				t.Fatalf("unexpected user: %+v", u)
			}
			if u.CreatedAt.IsZero() || u.UpdatedAt.IsZero() {
				t.Fatal("timestamps must be set")
			}
		})
	}
}

func TestUser_ChangeEmail(t *testing.T) {
	u, err := NewUser("a@b.com", "Alice", "hash")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	if err := u.ChangeEmail("bad"); !errors.Is(err, ErrInvalidEmail) {
		t.Fatalf("want ErrInvalidEmail, got %v", err)
	}
	if err := u.ChangeEmail("new@b.com"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.Email != "new@b.com" {
		t.Fatalf("email not changed: %s", u.Email)
	}
}

func TestUser_ChangeName(t *testing.T) {
	u, err := NewUser("a@b.com", "Alice", "hash")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	if err := u.ChangeName("  "); !errors.Is(err, ErrNameRequired) {
		t.Fatalf("want ErrNameRequired, got %v", err)
	}
	if err := u.ChangeName("Bob"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.Name != "Bob" {
		t.Fatalf("name not changed: %s", u.Name)
	}
}

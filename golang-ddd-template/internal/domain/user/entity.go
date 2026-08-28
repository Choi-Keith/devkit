// Package user contains the user aggregate: the entity, its domain
// invariants, and the persistence contract. The domain layer has zero
// third-party dependencies.
package user

import (
	"strings"
	"time"
)

// User is the aggregate root of the user domain.
type User struct {
	ID        uint64 // auto-increment, assigned by the database on Save
	Email     string
	Name      string
	Password  string // hashed password
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewUser creates a User after enforcing the domain invariants. The ID is
// assigned by the database (auto-increment) when the user is persisted.
func NewUser(email, name, passwordHash string) (*User, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return nil, ErrEmailRequired
	}
	if !strings.Contains(email, "@") {
		return nil, ErrInvalidEmail
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrNameRequired
	}
	if passwordHash == "" {
		return nil, ErrPasswordRequired
	}

	now := time.Now()
	return &User{
		Email:     email,
		Name:      name,
		Password:  passwordHash,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// ChangeEmail updates the email while keeping the domain invariant.
func (u *User) ChangeEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return ErrEmailRequired
	}
	if !strings.Contains(email, "@") {
		return ErrInvalidEmail
	}
	u.Email = email
	u.UpdatedAt = time.Now()
	return nil
}

// ChangeName updates the display name while keeping the domain invariant.
func (u *User) ChangeName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrNameRequired
	}
	u.Name = name
	u.UpdatedAt = time.Now()
	return nil
}

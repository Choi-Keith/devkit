package user

import "context"

// Repository defines the persistence contract for the user aggregate.
// Implementations live in the infra layer and are injected at runtime.
type Repository interface {
	// Save persists a new user and fills its auto-increment ID.
	Save(ctx context.Context, u *User) error
	// FindByID returns the user with the given id, or ErrNotFound.
	FindByID(ctx context.Context, id uint64) (*User, error)
	// FindByEmail returns the user with the given email, or ErrNotFound.
	FindByEmail(ctx context.Context, email string) (*User, error)
	// List returns a page of users and the total count.
	List(ctx context.Context, page, pageSize int) ([]*User, int64, error)
	// Delete removes the user with the given id, or returns ErrNotFound.
	Delete(ctx context.Context, id uint64) error
}

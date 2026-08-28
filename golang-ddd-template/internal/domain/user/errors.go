package user

import "errors"

// Domain errors are semantic sentinel errors. The HTTP adapter maps them
// to business error codes (ecode); the domain layer itself stays
// framework-agnostic.
var (
	ErrEmailRequired    = errors.New("email is required")
	ErrInvalidEmail     = errors.New("email is invalid")
	ErrNameRequired     = errors.New("name is required")
	ErrPasswordRequired = errors.New("password is required")
	ErrEmailTaken       = errors.New("email is already taken")
	ErrNotFound         = errors.New("user not found")
)

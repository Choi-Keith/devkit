package order

import "errors"

// Domain errors are semantic sentinel errors. The HTTP adapter maps them
// to business error codes (ecode); the domain layer itself stays
// framework-agnostic.
var (
	ErrUserIDRequired  = errors.New("user id is required")
	ErrOrderNoRequired = errors.New("order no is required")
	ErrInvalidAmount   = errors.New("amount must be positive")
	ErrNotFound        = errors.New("order not found")
	ErrNotCancellable  = errors.New("only pending orders can be cancelled")
)

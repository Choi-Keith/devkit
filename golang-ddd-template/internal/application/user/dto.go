// Package userapp orchestrates the user use cases: registration,
// queries and deletion. It depends only on domain contracts and stays
// free of transport or storage details.
package userapp

// RegisterInput carries the data needed to register a user.
type RegisterInput struct {
	Email    string
	Name     string
	Password string // plaintext, hashed before persistence
}

// UserDTO is the application-layer view of a user. The password hash is
// never exposed outside the domain.
type UserDTO struct {
	ID        uint64
	Email     string
	Name      string
	CreatedAt string
}

// UserListDTO is a page of users.
type UserListDTO struct {
	List     []UserDTO
	Total    int64
	Page     int
	PageSize int
}

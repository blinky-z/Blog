package models

// UserRole - represents user role
type UserRole string

// User - represents user without password. Use it to work with user when you don't need secret information
type User struct {
	Username string
	Email    string
	Role     UserRole
}

// LoginRequest - represents user login request
type LoginRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegistrationRequest - represents user registration request
type RegistrationRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

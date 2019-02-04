package models

// UserRole - represents user role
type UserRole string

// User - represents user without password. Use it to work with user when you don't need secret information as password
type User struct {
	Username string
	Email    string
	Role     UserRole
}

// LoginRequest - represents credentials that user inputs on login page.
type LoginRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegistrationRequest - represents credentials that user inputs on registration page.
type RegistrationRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

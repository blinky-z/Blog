package models

// UserRole - represents user role
type UserRole string

// User - represents user credentials
type User struct {
	Login    string `json:"login"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

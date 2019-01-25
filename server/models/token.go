package models

import "github.com/dgrijalva/jwt-go"

// TokenClaims - represents JWT token claims
type TokenClaims struct {
	Role string `json:"role"`
	jwt.StandardClaims
}
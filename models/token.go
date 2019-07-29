package models

import (
	"github.com/dgrijalva/jwt-go"
)

// TokenClaims - represents JWT token claims. It extends jwt.StandardClaims struct
type TokenClaims struct {
	Role        UserRole `json:"role"`
	Fingerprint string   `json:"fingerprint"`
	jwt.StandardClaims
}

package models

import (
	"context"
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const AuthPrincipalKey contextKey = "authPrincipal"

type AuthPrincipal struct {
	UserID int64  `json:"uid"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

type Claims struct {
	AuthPrincipal
	jwt.RegisteredClaims
}

// GetAuthPrincipal достаёт пользователя из контекста
func GetAuthPrincipal(ctx context.Context) (*AuthPrincipal, error) {
	val := ctx.Value(AuthPrincipalKey)
	principal, ok := val.(*AuthPrincipal)
	if !ok || principal == nil {
		return nil, errors.New("auth principal not found in context")
	}
	return principal, nil
}

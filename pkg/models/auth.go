package models

import "github.com/golang-jwt/jwt/v5"

type contextKey string

const UserIDKey contextKey = "userID"
const UsernameKey contextKey = "username"

type Claims struct {
	UserID int64  `json:"uid"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type AuthReqBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthQuery struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthDTO struct {
	Token string `json:"token"`
}

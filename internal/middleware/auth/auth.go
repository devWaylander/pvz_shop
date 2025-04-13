package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/devWaylander/pvz_store/api"
	internalErrors "github.com/devWaylander/pvz_store/pkg/errors"
	"github.com/devWaylander/pvz_store/pkg/log"
	"github.com/devWaylander/pvz_store/pkg/models"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	CreateUser(ctx context.Context, email, passwordHash, role string) (uuid.UUID, error)
	GetUserByEmail(ctx context.Context, email string) (*api.User, error)
	GetUserPassHashByUsername(ctx context.Context, email string) (string, error)
}

type middleware struct {
	repo   Repository
	jwtKey string
}

func NewMiddleware(repo Repository, jwtKey string) *middleware {
	return &middleware{
		repo:   repo,
		jwtKey: jwtKey,
	}
}

func (m *middleware) AuthContextEnrichingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if !strings.HasPrefix(authHeader, "Bearer ") {
			// это задача openapi3 AuthenticationFunc
			next.ServeHTTP(w, r)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.ParseWithClaims(tokenString, &models.Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(m.jwtKey), nil
		})
		if err != nil || !token.Valid {
			log.Logger.Err(errors.New(internalErrors.ErrInvalidToken)).Msg(err.Error())
			http.Error(w, internalErrors.ErrInvalidToken, http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(*models.Claims)
		if !ok {
			log.Logger.Err(errors.New(internalErrors.ErrInvalidClaims)).Msg("")
			http.Error(w, internalErrors.ErrInvalidClaims, http.StatusUnauthorized)
			return
		}

		ctx := models.SetAuthPrincipal(r.Context(), claims.AuthPrincipal)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (m *middleware) Middleware() openapi3filter.AuthenticationFunc {
	return func(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
		principal, err := models.GetAuthPrincipal(ctx)
		if err != nil || principal == nil {
			return fmt.Errorf(internalErrors.ErrUnauthenticated)
		}

		return nil
	}
}

func (m *middleware) DummyLogin(ctx context.Context, role api.UserRole) (api.Token, error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		log.Logger.Err(err).Msg("method DummyLogin, NewRandom")
		return api.Token(""), errors.New(internalErrors.ErrEncodeJWT)
	}
	claims := models.NewClaims(uuid, models.TestEmail, string(role))

	token, err := m.generateJWT(claims)
	if err != nil {
		log.Logger.Err(err).Msg("method DummyLogin, generateJWT")
		return api.Token(""), errors.New(internalErrors.ErrEncodeJWT)
	}

	return api.Token(token), nil
}

func (m *middleware) Registration(ctx context.Context, data api.PostRegisterJSONBody) (api.User, error) {
	ok := m.validatePassword(data.Password)
	if !ok {
		return api.User{}, errors.New(internalErrors.ErrWrongPasswordFormat)
	}

	password, err := m.passwordHash(data.Password)
	if err != nil {
		return api.User{}, err
	}

	uuid, err := m.repo.CreateUser(ctx, string(data.Email), password, string(data.Role))
	if err != nil {
		return api.User{}, err
	}

	user := api.User{
		Id:    &uuid,
		Email: data.Email,
		Role:  api.UserRole(data.Role),
	}

	return user, nil
}

func (m *middleware) Login(ctx context.Context, data api.PostLoginJSONBody) (api.Token, error) {
	user, err := m.repo.GetUserByEmail(ctx, string(data.Email))
	if err != nil {
		return api.Token(""), err
	}
	if user.Id == nil {
		return api.Token(""), errors.New(internalErrors.ErrUserNotFound)
	}

	passHash, err := m.repo.GetUserPassHashByUsername(ctx, string(data.Email))
	if err != nil {
		return api.Token(""), err
	}
	err = m.passwordCompare(data.Password, passHash)
	if err != nil {
		return api.Token(""), errors.New(internalErrors.ErrWrongPassword)
	}

	claims := models.NewClaims(*user.Id, string(user.Email), string(user.Role))

	token, err := m.generateJWT(claims)
	if err != nil {
		log.Logger.Err(err).Msg("method Login, generateJWT")
		return api.Token(""), errors.New(internalErrors.ErrEncodeJWT)
	}

	return token, nil
}

func (m *middleware) generateJWT(claims models.Claims) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	claims.RegisteredClaims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expirationTime),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(m.jwtKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (m *middleware) passwordHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (m *middleware) passwordCompare(password string, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return err
	}
	return nil
}

func (m *middleware) validatePassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`\d`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[\W_]`).MatchString(password)

	return hasUpper && hasLower && hasDigit && hasSpecial
}

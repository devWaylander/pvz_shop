package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/devWaylander/pvz_store/api"
	internalErrors "github.com/devWaylander/pvz_store/pkg/errors"
	"github.com/devWaylander/pvz_store/pkg/log"
	"github.com/devWaylander/pvz_store/pkg/models"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	// CreateUserTX(ctx context.Context, username, passwordHash string) (int64, error)
	// GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	// GetUserPassHashByUsername(ctx context.Context, username string) (string, error)
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

func (m *middleware) DummyLogin(ctx context.Context, role api.UserRole) api.Token {
	claims := models.NewClaims(9999, "test@test.com", string(role))

	token, err := m.generateJWT(claims)
	if err != nil {
		return api.Token("")
	}

	return api.Token(token)
}

// func (m *middleware) LoginWithPass(ctx context.Context, qp models.AuthQuery) (models.AuthDTO, error) {
// 	user, err := m.repo.GetUserByUsername(ctx, qp.Username)
// 	if err != nil {
// 		return models.AuthDTO{}, err
// 	}

// 	claims := models.Claims{
// 		UserID: user.ID,
// 		Email:  user.Email,
// 		Role:   user.Role,
// 	}

// 	// Не зарегистрирован
// 	if user.ID == 0 {
// 		validPass := m.validatePassword(qp.Password)
// 		if !validPass {
// 			return models.AuthDTO{}, errors.New(internalErrors.ErrWrongPasswordFormat)
// 		}

// 		validUsername := m.validateUsername(qp.Username)
// 		if !validUsername {
// 			return models.AuthDTO{}, errors.New(internalErrors.ErrWrongUsernameFormat)
// 		}

// 		passHash, err := m.passwordHash(qp.Password)
// 		if err != nil {
// 			return models.AuthDTO{}, err
// 		}
// 		userID, err := m.repo.CreateUserTX(ctx, qp.Username, passHash)
// 		if err != nil {
// 			return models.AuthDTO{}, err
// 		}

// 		claims.UserID = userID
// 		token, err := m.generateJWT(userID, qp.Username)
// 		if err != nil {
// 			return models.AuthDTO{}, err
// 		}

// 		return models.AuthDTO{Token: token}, err
// 	}

// 	passHash, err := m.repo.GetUserPassHashByUsername(ctx, qp.Username)
// 	if err != nil {
// 		return models.AuthDTO{}, err
// 	}
// 	err = m.passwordCompare(qp.Password, passHash)
// 	if err != nil {
// 		return models.AuthDTO{}, errors.New(internalErrors.ErrWrongPassword)
// 	}

// 	token, err := m.generateJWT(claims)
// 	if err != nil {
// 		return models.AuthDTO{}, err
// 	}

// 	return models.AuthDTO{Token: token}, nil
// }

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

func (m *middleware) validateUsername(username string) bool {
	if len(username) > 64 {
		return false
	}

	for _, r := range username {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}

	return true
}

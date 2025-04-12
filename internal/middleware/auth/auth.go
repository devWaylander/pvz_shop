package auth

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"

	internalErrors "github.com/devWaylander/coins_store/pkg/errors"
	"github.com/devWaylander/pvz_store/pkg/log"
	"github.com/devWaylander/pvz_store/pkg/models"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var unsecuredHandles = map[string]*struct{}{
	"/api/auth": {},
}

type Repository interface {
	CreateUserTX(ctx context.Context, username, passwordHash string) (int64, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	GetUserPassHashByUsername(ctx context.Context, username string) (string, error)
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

func (m *middleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if unsecuredHandles[r.URL.Path] != nil {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		ok := strings.Contains(authHeader, "Bearer ")
		if authHeader == "" || !ok {
			http.Error(w, internalErrors.ErrAuthHeader, http.StatusUnauthorized)
			return
		}

		tokenString := authHeader[len("Bearer "):]
		token, err := jwt.ParseWithClaims(tokenString, &models.Claims{}, func(token *jwt.Token) (any, error) {
			return []byte(m.jwtKey), nil
		})
		if err != nil {
			log.Logger.Err(err).Msg(err.Error())
			http.Error(w, internalErrors.ErrLogin, http.StatusUnauthorized)
			return
		}
		if !token.Valid {
			http.Error(w, internalErrors.ErrInvalidToken, http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(*models.Claims)
		if !ok {
			http.Error(w, internalErrors.ErrInvalidClaims, http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), models.UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, models.UsernameKey, claims.Username)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (m *middleware) LoginWithPass(ctx context.Context, qp models.AuthQuery) (models.AuthDTO, error) {
	user, err := m.repo.GetUserByUsername(ctx, qp.Username)
	if err != nil {
		return models.AuthDTO{}, err
	}

	claims := models.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
	}

	// Не зарегистрирован
	if user.ID == 0 {
		validPass := m.validatePassword(qp.Password)
		if !validPass {
			return models.AuthDTO{}, errors.New(internalErrors.ErrWrongPasswordFormat)
		}

		validUsername := m.validateUsername(qp.Username)
		if !validUsername {
			return models.AuthDTO{}, errors.New(internalErrors.ErrWrongUsernameFormat)
		}

		passHash, err := m.passwordHash(qp.Password)
		if err != nil {
			return models.AuthDTO{}, err
		}
		userID, err := m.repo.CreateUserTX(ctx, qp.Username, passHash)
		if err != nil {
			return models.AuthDTO{}, err
		}

		claims.UserID = userID
		token, err := m.generateJWT(userID, qp.Username)
		if err != nil {
			return models.AuthDTO{}, err
		}

		return models.AuthDTO{Token: token}, err
	}

	passHash, err := m.repo.GetUserPassHashByUsername(ctx, qp.Username)
	if err != nil {
		return models.AuthDTO{}, err
	}
	err = m.passwordCompare(qp.Password, passHash)
	if err != nil {
		return models.AuthDTO{}, errors.New(internalErrors.ErrWrongPassword)
	}

	token, err := m.generateJWT(claims)
	if err != nil {
		return models.AuthDTO{}, err
	}

	return models.AuthDTO{Token: token}, nil
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

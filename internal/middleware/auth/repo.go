package auth

import (
	"context"
	"database/sql"
	"errors"

	"github.com/devWaylander/pvz_store/api"
	internalErrors "github.com/devWaylander/pvz_store/pkg/errors"
	"github.com/devWaylander/pvz_store/pkg/log"
	"github.com/devWaylander/pvz_store/pkg/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type repository struct {
	db *sqlx.DB
}

func NewRepo(db *sqlx.DB) *repository {
	return &repository{db: db}
}

func (r *repository) CreateUser(ctx context.Context, email, passwordHash, role string) (uuid.UUID, error) {
	query := `
		INSERT INTO shop.users (email, password_hash, role)
		VALUES ($1, $2, $3)
		RETURNING id;
	`

	var uuid uuid.UUID
	err := r.db.GetContext(ctx, &uuid, query, email, passwordHash, role)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" && pqErr.Constraint == "users_email_key" {
				return uuid, errors.New(internalErrors.ErrUserExist)
			}
		}

		log.Logger.Err(err).Msg("method CreateUser")
		return uuid, errors.New("could not create user")
	}

	return uuid, nil
}

func (r *repository) GetUserByEmail(ctx context.Context, email string) (*api.User, error) {
	query := `SELECT id, email, role, password_hash, created_at FROM shop.users WHERE email = $1`

	var user models.UserDB

	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return &api.User{}, nil
		}

		log.Logger.Err(err).Msg("method GetUserByEmail")
		return nil, errors.New("could not get user")
	}

	return user.ToModelAPIUser(), nil
}

func (r *repository) GetUserPassHashByUsername(ctx context.Context, email string) (string, error) {
	query := `SELECT password_hash FROM shop.users WHERE email = $1`

	var passwordHash string

	err := r.db.GetContext(ctx, &passwordHash, query, email)
	if err != nil {
		log.Logger.Err(err).Msg("method GetUserPassHashByUsername")
		return "", errors.New("could not get password hash for user")
	}

	return passwordHash, nil
}

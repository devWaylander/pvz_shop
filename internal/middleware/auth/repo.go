package auth

import (
	"context"
	"errors"

	"github.com/devWaylander/pvz_store/pkg/log"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
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
		log.Logger.Err(err).Msg("method CreateUser")
		return uuid, errors.New("could not create user")
	}

	return uuid, nil
}

package repo

import (
	"context"
	"errors"
	"time"

	"github.com/devWaylander/pvz_store/api"
	internalErrors "github.com/devWaylander/pvz_store/pkg/errors"
	"github.com/devWaylander/pvz_store/pkg/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type repository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *repository {
	return &repository{db: db}
}

func (r *repository) CreatePVZ(ctx context.Context, id uuid.UUID, city string, registrationDate time.Time) (api.PVZ, error) {
	query := `
		INSERT INTO shop.pvz (id, city, registration_date)
		VALUES ($1, $2, $3)
		RETURNING id, city, registration_date
	`

	var inserted models.PvzDB
	err := r.db.QueryRowContext(ctx, query, id, city, registrationDate).
		Scan(&inserted.ID, &inserted.City, &inserted.RegistrationDate)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" && pqErr.Constraint == "pvz_pkey" {
			return api.PVZ{}, errors.New(internalErrors.ErrPVZExist)
		}

		return api.PVZ{}, err
	}

	return inserted.ToModelAPIPvz(), nil
}

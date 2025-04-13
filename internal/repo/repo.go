package repo

import (
	"context"
	"database/sql"
	"errors"
	"time"

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

		log.Logger.Err(err).Msg("method CreatePVZ")
		return api.PVZ{}, errors.New("could not create PVZ")
	}

	return inserted.ToModelAPIPvz(), nil
}

func (r *repository) IsPVZExist(ctx context.Context, id uuid.UUID) (bool, error) {
	query := `
        SELECT 1 FROM shop.pvz WHERE id = $1 LIMIT 1
    `

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}

		log.Logger.Err(err).Msg("method IsPVZExist")
		return false, errors.New("could not check if PVZ exists")
	}

	return true, nil
}

func (r *repository) CreateReception(ctx context.Context, pvzUUID uuid.UUID, status string) (api.Reception, error) {
	query := `
		INSERT INTO shop.receptions (pvz_id, status)
		VALUES ($1, $2)
		RETURNING id, pvz_id, created_at, status
	`

	var inserted models.ReceptionDB
	err := r.db.QueryRowContext(ctx, query, pvzUUID, status).
		Scan(&inserted.ID, &inserted.PvzID, &inserted.CreatedAt, &inserted.Status)

	if err != nil {
		log.Logger.Err(err).Msg("method CreateReception")
		return api.Reception{}, errors.New("could not create reception")
	}

	return inserted.ToModelAPIReception(), nil
}

func (r *repository) GetReceptionStatusByPvzUUID(ctx context.Context, pvzUUID uuid.UUID) (string, error) {
	query := `
		SELECT status
		FROM shop.receptions
		WHERE pvz_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var status string
	err := r.db.QueryRowContext(ctx, query, pvzUUID).Scan(&status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}

		log.Logger.Err(err).Msg("method GetReceptionStatusByPvzUUID")
		return "", errors.New("could not get reception status by pvz uuid")
	}

	return status, nil
}

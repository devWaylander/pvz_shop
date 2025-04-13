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

/*
PVZ
*/
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

/*
Reception
*/
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

func (r *repository) GetReceptionByPvzUUID(ctx context.Context, pvzUUID uuid.UUID) (api.Reception, error) {
	query := `
		SELECT id, pvz_id, status, created_at
		FROM shop.receptions
		WHERE pvz_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var reception models.ReceptionDB
	err := r.db.QueryRowContext(ctx, query, pvzUUID).
		Scan(&reception.ID, &reception.PvzID, &reception.Status, &reception.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return api.Reception{}, nil
		}
		log.Logger.Err(err).Str("pvz_uuid", pvzUUID.String()).Msg("method GetReceptionByPvzUUID")
		return api.Reception{}, errors.New("could not get reception by pvz uuid")
	}

	return reception.ToModelAPIReception(), nil
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

func (r *repository) UpdateReceptionStatus(ctx context.Context, recUUID uuid.UUID, status string) error {
	query := `
		UPDATE shop.receptions
		SET status = $1
		WHERE id = $2
	`

	_, err := r.db.ExecContext(ctx, query, status, recUUID)
	if err != nil {
		log.Logger.Err(err).
			Str("method", "UpdateReceptionStatus").
			Str("status", status).
			Str("reception_id", recUUID.String()).
			Msg("could not update reception status")

		return errors.New("could not update reception status")
	}

	return nil
}

/*
Product
*/
func (r *repository) CreateProduct(ctx context.Context, receptionUUID uuid.UUID, prType string) (api.Product, error) {
	query := `
		INSERT INTO shop.products (reception_id, type)
		VALUES ($1, $2)
		RETURNING id, reception_id, type, created_at
	`

	var inserted models.ProductDB
	err := r.db.QueryRowContext(ctx, query, receptionUUID, prType).
		Scan(&inserted.ID, &inserted.ReceptionID, &inserted.Type, &inserted.CreatedAt)

	if err != nil {
		log.Logger.Err(err).Str("reception_uuid", receptionUUID.String()).Str("type", prType).Msg("method CreateProduct")
		return api.Product{}, errors.New("could not create product")
	}

	return inserted.ToModelAPIProduct(), nil
}

func (r *repository) DeleteLastProductByReceptionUUID(ctx context.Context, receptionUUID uuid.UUID) error {
	query := `
		DELETE FROM shop.products
		WHERE id = (
			SELECT id
			FROM shop.products
			WHERE reception_id = $1
			ORDER BY created_at DESC
			LIMIT 1
		)
	`

	res, err := r.db.ExecContext(ctx, query, receptionUUID)
	if err != nil {
		log.Logger.Err(err).Str("receptionUUID", receptionUUID.String()).Msg("method DeleteLastProductByReceptionUUID")
		return errors.New("could not delete last product by reception uuid")
	}

	affected, err := res.RowsAffected()
	if err != nil {
		log.Logger.Err(err).
			Str("method", "DeleteLastProductByReceptionUUID").
			Str("receptionUUID", receptionUUID.String()).
			Msg("could not get rows affected")
		return errors.New("could not get deletion result")
	}

	if affected == 0 {
		return errors.New(internalErrors.ErrNoProductsToDelete)
	}

	return nil
}

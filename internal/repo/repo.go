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

func (r *repository) GetPVZs(ctx context.Context, page, limit int) ([]api.PVZ, error) {
	offset := (page - 1) * limit

	query := `
		SELECT id, city, registration_date
		FROM shop.pvz
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		log.Logger.Err(err).Msg("method GetPVZs")
		return nil, errors.New("could not get pvzs")
	}
	defer rows.Close()

	var pvzs []api.PVZ
	for rows.Next() {
		var pvz models.PvzDB
		if err := rows.Scan(&pvz.ID, &pvz.City, &pvz.RegistrationDate); err != nil {
			log.Logger.Err(err).Msg("method GetPVZs")
			return nil, errors.New("could not scan pvz row")
		}
		pvzs = append(pvzs, pvz.ToModelAPIPvz())
	}

	if err := rows.Err(); err != nil {
		log.Logger.Err(err).Msg("method GetPVZs")
		return nil, errors.New("error during rows iteration")
	}

	return pvzs, nil
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

func (r *repository) GetReceptionsByPvzUUIDsFiltered(ctx context.Context, pvzUUIDs []uuid.UUID, startDate, endDate *time.Time) ([]api.Reception, error) {
	query := `
		SELECT r.id, r.pvz_id, r.status, r.created_at
		FROM shop.receptions r
		WHERE r.pvz_id = ANY($1)
	`

	var args []any
	args = append(args, pq.Array(pvzUUIDs))

	if startDate != nil {
		query += ` AND r.created_at >= $2`
		args = append(args, *startDate)
	}
	if endDate != nil {
		query += ` AND r.created_at <= $3`
		args = append(args, *endDate)
	}

	query += ` ORDER BY r.created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		log.Logger.Err(err).Msg("method GetReceptionsByPvzUUIDsFiltered")
		return nil, errors.New("could not get receptions by pvz uuids")
	}
	defer rows.Close()

	var receptions []api.Reception
	for rows.Next() {
		var reception models.ReceptionDB
		if err := rows.Scan(&reception.ID, &reception.PvzID, &reception.Status, &reception.CreatedAt); err != nil {
			log.Logger.Err(err).Msg("method GetReceptionsByPvzUUIDsFiltered")
			return nil, errors.New("could not scan reception row")
		}
		receptions = append(receptions, reception.ToModelAPIReception())
	}

	if err := rows.Err(); err != nil {
		log.Logger.Err(err).Msg("method GetReceptionsByPvzUUIDsFiltered")
		return nil, errors.New("error during rows iteration")
	}

	return receptions, nil
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

func (r *repository) GetProductsByRecsUUIDs(ctx context.Context, recsUUIDs []uuid.UUID) ([]api.Product, error) {
	query := `
		SELECT p.id, p.reception_id, p.type, p.created_at
		FROM shop.products p
		WHERE p.reception_id = ANY($1)
	`

	rows, err := r.db.QueryContext(ctx, query, pq.Array(recsUUIDs))
	if err != nil {
		log.Logger.Err(err).Msg("method GetProductsByRecsUUIDs")
		return nil, errors.New("could not get products by reception uuids")
	}
	defer rows.Close()

	var products []api.Product
	for rows.Next() {
		var product models.ProductDB
		if err := rows.Scan(&product.ID, &product.ReceptionID, &product.Type, &product.CreatedAt); err != nil {
			log.Logger.Err(err).Msg("method GetProductsByRecsUUIDs")
			return nil, errors.New("could not scan product row")
		}
		products = append(products, product.ToModelAPIProduct())
	}

	if err := rows.Err(); err != nil {
		log.Logger.Err(err).Msg("method GetProductsByRecsUUIDs")
		return nil, errors.New("error during rows iteration")
	}

	return products, nil
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

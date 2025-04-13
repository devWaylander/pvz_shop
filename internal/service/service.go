package service

import (
	"context"
	"errors"
	"time"

	"github.com/devWaylander/pvz_store/api"
	internalErrors "github.com/devWaylander/pvz_store/pkg/errors"
	"github.com/devWaylander/pvz_store/pkg/models"
	"github.com/google/uuid"
)

type Repository interface {
	// PVZ
	CreatePVZ(ctx context.Context, id uuid.UUID, city string, registrationDate time.Time) (api.PVZ, error)
	IsPVZExist(ctx context.Context, id uuid.UUID) (bool, error)
	GetPVZsWithPagination(ctx context.Context, page, limit int) ([]api.PVZ, error)
	// Reception
	CreateReception(ctx context.Context, pvzUUID uuid.UUID, status string) (api.Reception, error)
	GetReceptionByPvzUUID(ctx context.Context, pvzUUID uuid.UUID) (api.Reception, error)
	GetReceptionsByPvzUUIDsFiltered(ctx context.Context, pvzUUIDs []uuid.UUID, startDate, endDate *time.Time) ([]api.Reception, error)
	GetReceptionStatusByPvzUUID(ctx context.Context, pvzUUID uuid.UUID) (string, error)
	UpdateReceptionStatus(ctx context.Context, recUUID uuid.UUID, status string) error
	// Product
	CreateProduct(ctx context.Context, receptionUUID uuid.UUID, prType string) (api.Product, error)
	GetProductsByRecsUUIDs(ctx context.Context, recsUUIDs []uuid.UUID) ([]api.Product, error)
	DeleteLastProductByReceptionUUID(ctx context.Context, receptionUUID uuid.UUID) error
}

type service struct {
	repo Repository
}

func New(repo Repository) *service {
	return &service{
		repo: repo,
	}
}

/*
PVZ
*/
func (s *service) CreatePVZ(ctx context.Context, data api.PVZ) (api.PVZ, error) {
	if data.RegistrationDate != nil && data.RegistrationDate.After(time.Now()) {
		return api.PVZ{}, errors.New(internalErrors.ErrWrongRegDate)
	}

	pvz, err := s.repo.CreatePVZ(ctx, *data.Id, string(data.City), *data.RegistrationDate)
	if err != nil {
		return api.PVZ{}, err
	}

	return pvz, nil
}

func (s *service) GetPVZsInfo(ctx context.Context, data api.GetPvzParams) ([]models.PvzInfo, error) {
	if data.Page == nil || *data.Page < 1 {
		page := 1
		data.Page = &page
	}
	if data.Limit == nil || *data.Limit < 1 {
		limit := 10
		data.Limit = &limit
	}

	pvzs, err := s.repo.GetPVZsWithPagination(ctx, *data.Page, *data.Limit)
	if err != nil {
		return nil, err
	}

	pvzsUUIDs := make([]uuid.UUID, 0, len(pvzs))
	for _, pvz := range pvzs {
		if pvz.Id == nil {
			return nil, errors.New(internalErrors.ErrPVZDoesntExist)
		}
		pvzsUUIDs = append(pvzsUUIDs, *pvz.Id)
	}

	receptions, err := s.repo.GetReceptionsByPvzUUIDsFiltered(ctx, pvzsUUIDs, data.StartDate, data.EndDate)
	if err != nil {
		return nil, err
	}

	recsUUIDs := make([]uuid.UUID, 0, len(receptions))
	for _, rec := range receptions {
		if rec.Id == nil {
			return nil, errors.New(internalErrors.ErrReceptionDoesntExist)
		}
		recsUUIDs = append(recsUUIDs, *rec.Id)
	}

	products, err := s.repo.GetProductsByRecsUUIDs(ctx, recsUUIDs)
	if err != nil {
		return nil, err
	}

	// группируем продукты по reception_id
	productsByRec := make(map[uuid.UUID][]api.Product)
	for _, product := range products {
		productsByRec[product.ReceptionId] = append(productsByRec[product.ReceptionId], product)
	}

	// группируем приёмки по pvz_id
	recsByPvz := make(map[uuid.UUID][]models.ReceptionWithProducts)
	for _, rec := range receptions {
		recsByPvz[rec.PvzId] = append(recsByPvz[rec.PvzId], models.ReceptionWithProducts{
			Reception: rec,
			// проверено на nil ранее
			Products: productsByRec[*rec.Id],
		})
	}

	// собираем финальный список PvzInfo
	var result []models.PvzInfo
	for _, pvz := range pvzs {
		if pvz.Id == nil {
			continue
		}
		result = append(result, models.PvzInfo{
			Pvz: pvz,
			// проверено на nil ранее
			Receptions: recsByPvz[*pvz.Id],
		})
	}

	return result, nil
}

/*
Reception
*/
func (s *service) CreateReception(ctx context.Context, data api.PostReceptionsJSONBody) (api.Reception, error) {
	isPVZExist, err := s.repo.IsPVZExist(ctx, data.PvzId)
	if err != nil {
		return api.Reception{}, err
	}
	if !isPVZExist {
		return api.Reception{}, errors.New(internalErrors.ErrPVZDoesntExist)
	}

	pvzStatus, err := s.repo.GetReceptionStatusByPvzUUID(ctx, data.PvzId)
	if err != nil {
		return api.Reception{}, err
	}
	if pvzStatus == string(api.InProgress) {
		return api.Reception{}, errors.New(internalErrors.ErrReceptionExist)
	}

	reception, err := s.repo.CreateReception(ctx, data.PvzId, string(api.InProgress))
	if err != nil {
		return api.Reception{}, err
	}

	return reception, nil
}

func (s *service) CloseReception(ctx context.Context, pvzUUID uuid.UUID) (api.Reception, error) {
	rec, err := s.getReceptionByPvzUUID(ctx, pvzUUID)
	if err != nil {
		return api.Reception{}, err
	}

	err = s.repo.UpdateReceptionStatus(ctx, *rec.Id, string(api.Close))
	if err != nil {
		return api.Reception{}, err
	}
	rec.Status = api.Close

	return rec, nil
}

/*
Product
*/
func (s *service) CreateProduct(ctx context.Context, data api.PostProductsJSONBody) (api.Product, error) {
	rec, err := s.getReceptionByPvzUUID(ctx, data.PvzId)
	if err != nil {
		return api.Product{}, err
	}

	product, err := s.repo.CreateProduct(ctx, *rec.Id, string(data.Type))
	if err != nil {
		return api.Product{}, err
	}

	return product, nil
}

func (s *service) DeleteLastProduct(ctx context.Context, pvzUUID uuid.UUID) error {
	rec, err := s.getReceptionByPvzUUID(ctx, pvzUUID)
	if err != nil {
		return err
	}

	return s.repo.DeleteLastProductByReceptionUUID(ctx, *rec.Id)
}

func (s *service) getReceptionByPvzUUID(ctx context.Context, pvzUUID uuid.UUID) (api.Reception, error) {
	isPVZExist, err := s.repo.IsPVZExist(ctx, pvzUUID)
	if err != nil {
		return api.Reception{}, err
	}
	if !isPVZExist {
		return api.Reception{}, errors.New(internalErrors.ErrPVZDoesntExist)
	}

	reception, err := s.repo.GetReceptionByPvzUUID(ctx, pvzUUID)
	if err != nil {
		return api.Reception{}, err
	}
	if reception.Id == nil {
		return api.Reception{}, errors.New(internalErrors.ErrReceptionDoesntExist)
	}
	if reception.Status != api.InProgress {
		return api.Reception{}, errors.New(internalErrors.ErrWrongReceptionStatus)
	}

	return reception, nil
}

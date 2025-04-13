package service

import (
	"context"
	"errors"
	"time"

	"github.com/devWaylander/pvz_store/api"
	internalErrors "github.com/devWaylander/pvz_store/pkg/errors"
	"github.com/google/uuid"
)

type Repository interface {
	// PVZ
	CreatePVZ(ctx context.Context, id uuid.UUID, city string, registrationDate time.Time) (api.PVZ, error)
	IsPVZExist(ctx context.Context, id uuid.UUID) (bool, error)
	// Reception
	CreateReception(ctx context.Context, pvzUUID uuid.UUID, status string) (api.Reception, error)
	GetReceptionByPvzUUID(ctx context.Context, pvzUUID uuid.UUID) (api.Reception, error)
	GetReceptionStatusByPvzUUID(ctx context.Context, pvzUUID uuid.UUID) (string, error)
	// Product
	CreateProduct(ctx context.Context, receptionUUID uuid.UUID, prType string) (api.Product, error)
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

/*
Product
*/
func (s *service) CreateProduct(ctx context.Context, data api.PostProductsJSONBody) (api.Product, error) {
	recUUID, err := s.validateReceptionForProduct(ctx, data.PvzId)
	if err != nil {
		return api.Product{}, err
	}

	product, err := s.repo.CreateProduct(ctx, recUUID, string(data.Type))
	if err != nil {
		return api.Product{}, err
	}

	return product, nil
}

func (s *service) DeleteLastProduct(ctx context.Context, pvzUUID uuid.UUID) error {
	recUUID, err := s.validateReceptionForProduct(ctx, pvzUUID)
	if err != nil {
		return err
	}

	return s.repo.DeleteLastProductByReceptionUUID(ctx, recUUID)
}

func (s *service) validateReceptionForProduct(ctx context.Context, pvzUUID uuid.UUID) (uuid.UUID, error) {
	isPVZExist, err := s.repo.IsPVZExist(ctx, pvzUUID)
	if err != nil {
		return uuid.UUID{}, err
	}
	if !isPVZExist {
		return uuid.UUID{}, errors.New(internalErrors.ErrPVZDoesntExist)
	}

	reception, err := s.repo.GetReceptionByPvzUUID(ctx, pvzUUID)
	if err != nil {
		return uuid.UUID{}, err
	}
	if reception.Id == nil {
		return uuid.UUID{}, errors.New(internalErrors.ErrReceptionDoesntExist)
	}
	if reception.Status != api.InProgress {
		return uuid.UUID{}, errors.New(internalErrors.ErrWrongReceptionStatus)
	}

	return *reception.Id, nil
}

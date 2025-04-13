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
	GetReceptionStatusByPvzUUID(ctx context.Context, pvzUUID uuid.UUID) (string, error)
}

type service struct {
	repo Repository
}

func New(repo Repository) *service {
	return &service{
		repo: repo,
	}
}

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

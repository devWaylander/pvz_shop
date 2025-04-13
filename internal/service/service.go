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
	CreatePVZ(ctx context.Context, id uuid.UUID, city string, registrationDate time.Time) (api.PVZ, error)
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

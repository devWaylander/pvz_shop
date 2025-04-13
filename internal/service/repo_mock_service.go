package service

import (
	"context"
	"time"

	"github.com/devWaylander/pvz_store/api"
	"github.com/google/uuid"
)

type MockRepository struct {
	// PVZ
	CreatePVZFunc  func(ctx context.Context, id uuid.UUID, city string, registrationDate time.Time) (api.PVZ, error)
	IsPVZExistFunc func(ctx context.Context, id uuid.UUID) (bool, error)
	GetPVZsFunc    func(ctx context.Context, page, limit int) ([]api.PVZ, error)
	// Reception
	CreateReceptionFunc                 func(ctx context.Context, pvzUUID uuid.UUID, status string) (api.Reception, error)
	GetReceptionByPvzUUIDFunc           func(ctx context.Context, pvzUUID uuid.UUID) (api.Reception, error)
	GetReceptionsByPvzUUIDsFilteredFunc func(ctx context.Context, pvzUUIDs []uuid.UUID, startDate, endDate *time.Time) ([]api.Reception, error)
	GetReceptionStatusByPvzUUIDFunc     func(ctx context.Context, pvzUUID uuid.UUID) (string, error)
	UpdateReceptionStatusFunc           func(ctx context.Context, recUUID uuid.UUID, status string) error
	// Product
	CreateProductFunc                    func(ctx context.Context, receptionUUID uuid.UUID, prType string) (api.Product, error)
	GetProductsByRecsUUIDsFunc           func(ctx context.Context, recsUUIDs []uuid.UUID) ([]api.Product, error)
	DeleteLastProductByReceptionUUIDFunc func(ctx context.Context, receptionUUID uuid.UUID) error
}

func (m *MockRepository) CreatePVZ(ctx context.Context, id uuid.UUID, city string, registrationDate time.Time) (api.PVZ, error) {
	return m.CreatePVZFunc(ctx, id, city, registrationDate)
}

func (m *MockRepository) IsPVZExist(ctx context.Context, id uuid.UUID) (bool, error) {
	return m.IsPVZExistFunc(ctx, id)
}

func (m *MockRepository) GetPVZs(ctx context.Context, page, limit int) ([]api.PVZ, error) {
	return m.GetPVZsFunc(ctx, page, limit)
}

func (m *MockRepository) CreateReception(ctx context.Context, pvzUUID uuid.UUID, status string) (api.Reception, error) {
	return m.CreateReceptionFunc(ctx, pvzUUID, status)
}

func (m *MockRepository) GetReceptionByPvzUUID(ctx context.Context, pvzUUID uuid.UUID) (api.Reception, error) {
	return m.GetReceptionByPvzUUIDFunc(ctx, pvzUUID)
}

func (m *MockRepository) GetReceptionsByPvzUUIDsFiltered(ctx context.Context, pvzUUIDs []uuid.UUID, startDate, endDate *time.Time) ([]api.Reception, error) {
	return m.GetReceptionsByPvzUUIDsFilteredFunc(ctx, pvzUUIDs, startDate, endDate)
}

func (m *MockRepository) GetReceptionStatusByPvzUUID(ctx context.Context, pvzUUID uuid.UUID) (string, error) {
	return m.GetReceptionStatusByPvzUUIDFunc(ctx, pvzUUID)
}

func (m *MockRepository) UpdateReceptionStatus(ctx context.Context, recUUID uuid.UUID, status string) error {
	return m.UpdateReceptionStatusFunc(ctx, recUUID, status)
}

func (m *MockRepository) CreateProduct(ctx context.Context, receptionUUID uuid.UUID, prType string) (api.Product, error) {
	return m.CreateProductFunc(ctx, receptionUUID, prType)
}

func (m *MockRepository) GetProductsByRecsUUIDs(ctx context.Context, recsUUIDs []uuid.UUID) ([]api.Product, error) {
	return m.GetProductsByRecsUUIDsFunc(ctx, recsUUIDs)
}

func (m *MockRepository) DeleteLastProductByReceptionUUID(ctx context.Context, receptionUUID uuid.UUID) error {
	return m.DeleteLastProductByReceptionUUIDFunc(ctx, receptionUUID)
}

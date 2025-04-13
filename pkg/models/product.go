package models

import (
	"time"

	"github.com/devWaylander/pvz_store/api"
	"github.com/go-openapi/strfmt"
	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
)

type ProductDB struct {
	ID          uuid.UUID       `db:"id"`
	Type        string          `db:"type"`
	ReceptionID uuid.UUID       `db:"reception_id"`
	CreatedAt   strfmt.DateTime `db:"created_at"`
}

func (pdb *ProductDB) ToModelAPIProduct() api.Product {
	id := types.UUID(pdb.ID)
	receptionId := types.UUID(pdb.ReceptionID)
	return api.Product{
		Id:          &id,
		Type:        api.ProductType(pdb.Type),
		ReceptionId: receptionId,
		DateTime:    (*time.Time)(&pdb.CreatedAt),
	}
}

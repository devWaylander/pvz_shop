package models

import (
	"time"

	"github.com/devWaylander/pvz_store/api"
	"github.com/go-openapi/strfmt"
	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
)

type PvzDB struct {
	ID               uuid.UUID       `db:"id"`
	City             string          `db:"city"`
	RegistrationDate strfmt.DateTime `db:"registration_date"`
}

func (pvzdb *PvzDB) ToModelAPIPvz() api.PVZ {
	id := types.UUID(pvzdb.ID)
	return api.PVZ{
		Id:               &id,
		City:             api.PVZCity(pvzdb.City),
		RegistrationDate: (*time.Time)(&pvzdb.RegistrationDate),
	}
}

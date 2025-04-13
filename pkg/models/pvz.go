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

type PvzInfoDB struct {
	PvzID               uuid.UUID       `db:"pvz_id"`
	PvzCity             string          `db:"city"`
	PvzRegistrationDate strfmt.DateTime `db:"registration_date"`
	ReceptionID         uuid.UUID       `db:"id"`
	ReceptionPvzID      uuid.UUID       `db:"pvz_id"`
	ReceptionStatus     string          `db:"status"`
	ReceptionCreatedAt  strfmt.DateTime `db:"created_at"`
}

type PvzInfo struct {
	Pvz        api.PVZ
	Receptions []ReceptionWithProducts
}

type ReceptionWithProducts struct {
	Reception api.Reception
	Products  []api.Product
}

func MapPvzInfoToAPIResponse(pvzsInfo []PvzInfo) api.GetPvz200JSONResponse {
	var response api.GetPvz200JSONResponse
	for _, pvzInfo := range pvzsInfo {
		var receptions []struct {
			Products  *[]api.Product `json:"products,omitempty"`
			Reception *api.Reception `json:"reception,omitempty"`
		}

		for _, rwp := range pvzInfo.Receptions {
			products := rwp.Products
			receptions = append(receptions, struct {
				Products  *[]api.Product `json:"products,omitempty"`
				Reception *api.Reception `json:"reception,omitempty"`
			}{
				Products:  &products,
				Reception: &rwp.Reception,
			})
		}

		response = append(response, struct {
			Pvz        *api.PVZ `json:"pvz,omitempty"`
			Receptions *[]struct {
				Products  *[]api.Product `json:"products,omitempty"`
				Reception *api.Reception `json:"reception,omitempty"`
			} `json:"receptions,omitempty"`
		}{
			Pvz:        &pvzInfo.Pvz,
			Receptions: &receptions,
		})
	}

	return response
}

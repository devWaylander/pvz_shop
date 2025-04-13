package models

import (
	"time"

	"github.com/devWaylander/pvz_store/api"
	"github.com/go-openapi/strfmt"
	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
)

type ReceptionDB struct {
	ID        uuid.UUID       `db:"id"`
	PvzID     uuid.UUID       `db:"pvz_id"`
	Status    string          `db:"status"`
	CreatedAt strfmt.DateTime `db:"created_at"`
}

func (rdb *ReceptionDB) ToModelAPIReception() api.Reception {
	id := types.UUID(rdb.ID)
	pvzId := types.UUID(rdb.PvzID)
	return api.Reception{
		Id:       &id,
		PvzId:    pvzId,
		Status:   api.ReceptionStatus(rdb.Status),
		DateTime: time.Time(rdb.CreatedAt),
	}
}

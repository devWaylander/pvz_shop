package models

import (
	"github.com/devWaylander/pvz_store/api"
	"github.com/go-openapi/strfmt"
	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
)

type UserDB struct {
	ID           uuid.UUID       `db:"id"`
	Email        string          `db:"email"`
	Role         string          `db:"role"`
	PasswordHash string          `db:"password_hash"`
	CreatedAt    strfmt.DateTime `db:"created_at"`
}

func (udb *UserDB) ToModelAPIUser() *api.User {
	id := types.UUID(udb.ID)
	return &api.User{
		Id:    &id,
		Email: types.Email(udb.Email),
		Role:  api.UserRole(udb.Role),
	}
}

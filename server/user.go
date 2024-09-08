package main

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type User struct {
	ID uuid.UUID `db:"id"`
}

func (u *User) GetHibpCases(ctx context.Context, tx pgx.Tx) ([]HibpCase, error) {
	return RetrieveHibpCasesByUserId(ctx, tx, u.ID)
}

package main

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type HibpCase struct {
	ID     string    `db:"id"`
	UserId uuid.UUID `db:"user_id"`
	Name   string    `db:"name"`
}

func RetrieveHibpCasesByUserId(ctx context.Context, tx pgx.Tx, id uuid.UUID) ([]HibpCase, error) {
	rows, _ := tx.Query(ctx, "SELECT * FROM hibp_cases WHERE user_id = $1", id)
	return pgx.CollectRows(rows, pgx.RowToStructByName[HibpCase])
}

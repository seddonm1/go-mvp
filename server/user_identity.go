package main

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type IdentityUser struct {
	ID     string    `db:"id"`
	UserId uuid.UUID `db:"user_id"`
}

func RetrieveIdentityUserById(ctx context.Context, tx pgx.Tx, id string) (*IdentityUser, error) {
	rows, _ := tx.Query(ctx, "SELECT * FROM identity_users WHERE id = $1", id)
	return pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[IdentityUser])
}

func (i *IdentityUser) RetrieveUser(ctx context.Context, tx pgx.Tx) (*User, error) {
	rows, _ := tx.Query(ctx, "SELECT * FROM users WHERE id = $1", i.UserId)
	return pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[User])
}

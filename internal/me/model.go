package me

import (
	"context"

	"github.com/acoshift/pgsql/pgctx"
)

func setUserPhoto(ctx context.Context, userID string, photo string) error {
	// language=SQL
	_, err := pgctx.Exec(ctx, `
		update users
		set photo = $2
		where id = $1
	`, userID, photo)
	return err
}

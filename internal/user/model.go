package user

import (
	"context"
	"database/sql"
	"errors"

	"github.com/acoshift/pgsql/pgctx"
)

var (
	errNotFound = errors.New("user: not found")
)

func getUserIDFromUsername(ctx context.Context, username string) (userID string, err error) {
	// language=SQL
	err = pgctx.QueryRow(ctx, `
		select id
		from users
		where username = $1
	`, username).Scan(&userID)
	if err == sql.ErrNoRows {
		return "", errNotFound
	}
	return
}

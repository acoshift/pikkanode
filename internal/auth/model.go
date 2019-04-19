package auth

import (
	"context"
	"database/sql"
	"errors"

	"github.com/acoshift/pgsql"
	"github.com/acoshift/pgsql/pgctx"
)

var (
	errUsernameDuplicated = errors.New("auth: username duplicated")
	errNotFound           = errors.New("auth: not found")
)

func insertUser(ctx context.Context, username, hashedPassword string) (userID string, err error) {
	// language=SQL
	err = pgctx.QueryRow(ctx, `
		insert into users
			(username, password)
		values
			($1, $2)
		returning id
	`, username, hashedPassword).Scan(&userID)
	if pgsql.IsUniqueViolation(err, "users_username_idx") {
		return "", errUsernameDuplicated
	}
	return
}

func getUserIDAndPasswordByUsername(ctx context.Context, username string) (userID, hashedPassword string, err error) {
	// language=SQL
	err = pgctx.QueryRow(ctx, `
		select id, password
		from users
		where username = $1
	`, username).Scan(&userID, &hashedPassword)
	if err == sql.ErrNoRows {
		return "", "", errNotFound
	}
	return
}

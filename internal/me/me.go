package me

import (
	"context"
	"database/sql"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/acoshift/pgsql/pgctx"

	"github.com/acoshift/pikkanode/internal/session"
)

type ProfileRequest struct{}

func (*ProfileRequest) AdaptRequest(r *http.Request) {
	if r.Method == http.MethodGet {
		r.Method = http.MethodPost
		r.Header.Set("Content-Type", "application/json")
		r.Body = ioutil.NopCloser(strings.NewReader("{}"))
	}
}

type ProfileResult struct {
	Username string `json:"username"`
}

func Profile(ctx context.Context, _ *ProfileRequest) (*ProfileResult, error) {
	userID := session.GetUserID(ctx)
	if userID == "" {
		return nil, errInvalidCredentials
	}

	var r ProfileResult
	// language=SQL
	err := pgctx.QueryRow(ctx, `
		select username
		from users
		where id = $1
	`, userID).Scan(
		&r.Username,
	)
	if err == sql.ErrNoRows {
		// user removed ?
		session.Get(ctx).Destroy()
		return nil, errInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	return &r, nil
}

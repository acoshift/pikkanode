package user

import (
	"context"
	"database/sql"
	"net/http"
	"net/url"
	"unicode/utf8"

	"github.com/acoshift/pgsql/pgctx"

	"github.com/acoshift/pikkanode/internal/validator"
)

type ProfileRequest struct {
	Username string `json:"username"`
}

func (*ProfileRequest) AdaptRequest(r *http.Request) {
	if r.Method == http.MethodGet {
		r.ParseForm()
		r.Method = http.MethodPost
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.PostForm = r.Form
	}
}

func (req *ProfileRequest) UnmarshalForm(v url.Values) error {
	req.Username = v.Get("username")
	return nil
}

func (req *ProfileRequest) Valid() error {
	v := validator.New()
	v.Must(req.Username != "", "username required")
	v.Must(utf8.RuneCountInString(req.Username) <= 15, "username maximum 15 characters")

	return v.Error()
}

type ProfileResult struct {
	Username string `json:"username"`
}

func Profile(ctx context.Context, req *ProfileRequest) (*ProfileResult, error) {
	var r ProfileResult
	// language=SQL
	err := pgctx.QueryRow(ctx, `
		select username
		from users
		where username = $1
	`, req.Username).Scan(
		&r.Username,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &r, nil
}

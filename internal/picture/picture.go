package picture

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/acoshift/pgsql/pgctx"
	"github.com/lib/pq"

	"github.com/acoshift/pikkanode/internal/validator"
)

type GetRequest struct {
	ID string `json:"id"`
}

func (*GetRequest) AdaptRequest(r *http.Request) {
	if r.Method == http.MethodGet {
		r.ParseForm()
		r.Method = http.MethodPost
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.PostForm = r.Form
	}
}

func (req *GetRequest) UnmarshalForm(v url.Values) error {
	req.ID = v.Get("id")
	return nil
}

func (req *GetRequest) Valid() error {
	v := validator.New()
	v.Must(req.ID != "", "id required")
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		v.Add(errors.New("invalid id"))
	} else {
		v.Must(id > 0, "id required")
	}

	return v.Error()
}

type GetResult struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Detail    string    `json:"detail"`
	Photo     string    `json:"photo"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"createdAt"`
}

func Get(ctx context.Context, req *GetRequest) (*GetResult, error) {
	var r GetResult
	// language=SQL
	err := pgctx.QueryRow(ctx, `
		select
			id, name, detail, photo, tags, created_at
		from pictures
		where id = $1
	`, req.ID).Scan(
		&r.ID, &r.Name, &r.Detail, &r.Photo, pq.Array(&r.Tags), &r.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, errPictureNotFound
	}
	if err != nil {
		return nil, err
	}

	return &r, nil
}

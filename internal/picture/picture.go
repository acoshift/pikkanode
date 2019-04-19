package picture

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/acoshift/pgsql"
	"github.com/acoshift/pgsql/pgctx"
	"github.com/lib/pq"

	"github.com/acoshift/pikkanode/internal/session"
	"github.com/acoshift/pikkanode/internal/validator"
)

type GetRequest struct {
	ID string `json:"id"`
}

func (req *GetRequest) Valid() error {
	v := validator.New()
	v.Must(req.ID != "", "id required")
	if req.ID != "" {
		id, err := strconv.ParseInt(req.ID, 10, 64)
		if err != nil {
			v.Add(errors.New("invalid id"))
		} else {
			v.Must(id > 0, "id required")
		}
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

type FavoriteRequest struct {
	ID string `json:"id"`
}

func (req *FavoriteRequest) Valid() error {
	v := validator.New()
	v.Must(req.ID != "", "id required")
	if req.ID != "" {
		id, err := strconv.ParseInt(req.ID, 10, 64)
		if err != nil {
			v.Add(errors.New("invalid id"))
		} else {
			v.Must(id > 0, "id required")
		}
	}

	return v.Error()
}

func Favorite(ctx context.Context, req *FavoriteRequest) (*struct{}, error) {
	userID := session.GetUserID(ctx)
	if userID == "" {
		return nil, errInvalidCredentials
	}

	var r struct{}
	// language=SQL
	_, err := pgctx.Exec(ctx, `
		insert into favorites
			(user_id, picture_id)
		values
			($1, $2)
		on conflict do nothing
	`, userID, req.ID)
	if pgsql.IsForeignKeyViolation(err, "favorites_picture_id_fkey") {
		return nil, errPictureNotFound
	}
	if err != nil {
		return nil, err
	}

	return &r, nil
}

type UnFavoriteRequest struct {
	ID string `json:"id"`
}

func (req *UnFavoriteRequest) Valid() error {
	v := validator.New()
	v.Must(req.ID != "", "id required")
	if req.ID != "" {
		id, err := strconv.ParseInt(req.ID, 10, 64)
		if err != nil {
			v.Add(errors.New("invalid id"))
		} else {
			v.Must(id > 0, "id required")
		}
	}

	return v.Error()
}

func UnFavorite(ctx context.Context, req *UnFavoriteRequest) (*struct{}, error) {
	userID := session.GetUserID(ctx)
	if userID == "" {
		return nil, errInvalidCredentials
	}

	var r struct{}
	// language=SQL
	_, err := pgctx.Exec(ctx, `
		delete from favorites where user_id = $1 and picture_id = $2
	`, userID, req.ID)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

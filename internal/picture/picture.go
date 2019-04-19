package picture

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"
	"unicode/utf8"

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

type CommentItem struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
	User      struct {
		Username string `json:"username"`
		Photo    string `json:"photo"`
	} `json:"user"`
}

type GetResult struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Detail    string         `json:"detail"`
	Photo     string         `json:"photo"`
	Tags      []string       `json:"tags"`
	Comments  []*CommentItem `json:"comments"`
	CreatedAt time.Time      `json:"createdAt"`
}

func Get(ctx context.Context, req *GetRequest) (*GetResult, error) {
	var r GetResult

	{
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

	}

	{
		// language=SQL
		rows, err := pgctx.Query(ctx, `
			select
				c.id, c.content, c.created_at,
				u.username, u.photo
			from comments c
				left join users u on c.user_id = u.id
			where c.picture_id = $1
		`, req.ID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		r.Comments = make([]*CommentItem, 0)
		for rows.Next() {
			var x CommentItem
			err := rows.Scan(
				&x.ID, &x.Content, &x.CreatedAt,
				&x.User.Username, &x.User.Photo,
			)
			if err != nil {
				return nil, err
			}
			r.Comments = append(r.Comments, &x)
		}

		if err := rows.Err(); err != nil {
			return nil, err
		}
	}

	return &r, nil
}

type FavoriteRequest struct {
	ID       string `json:"id"`
	Favorite bool   `json:"favorite"`
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

	if req.Favorite {
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
	} else {
		// language=SQL
		_, err := pgctx.Exec(ctx, `
			delete from favorites where user_id = $1 and picture_id = $2
		`, userID, req.ID)
		if err != nil {
			return nil, err
		}
	}

	return &r, nil
}

type CommentRequest struct {
	PictureID string `json:"pictureId"`
	Content   string `json:"content"`
}

func (req *CommentRequest) Valid() error {
	v := validator.New()
	v.Must(req.PictureID != "", "pictureId required")
	if req.PictureID != "" {
		id, err := strconv.ParseInt(req.PictureID, 10, 64)
		if err != nil {
			v.Add(errors.New("invalid picture id"))
		} else {
			v.Must(id > 0, "pictureId required")
		}
	}
	v.Must(req.Content != "", "content required")
	v.Must(utf8.RuneCountInString(req.Content) <= 255, "content length maximum 255 charactors")

	return v.Error()
}

func Comment(ctx context.Context, req *CommentRequest) (*struct{}, error) {
	userID := session.GetUserID(ctx)
	if userID == "" {
		return nil, errInvalidCredentials
	}

	var r struct{}
	// language=SQL
	_, err := pgctx.Exec(ctx, `
		insert into comments
			(user_id, picture_id, content)
		values
			($1, $2, $3)
	`, userID, req.PictureID, req.Content)
	if pgsql.IsForeignKeyViolation(err, "comments_picture_id_fkey") {
		return nil, errPictureNotFound
	}
	if err != nil {
		return nil, err
	}

	return &r, nil
}

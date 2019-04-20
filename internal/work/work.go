package work

import (
	"context"
	"database/sql"
	"time"
	"unicode/utf8"

	"github.com/acoshift/pgsql"
	"github.com/acoshift/pgsql/pgctx"
	"github.com/asaskevich/govalidator"
	"github.com/lib/pq"

	"github.com/acoshift/pikkanode/internal/file"
	"github.com/acoshift/pikkanode/internal/session"
	"github.com/acoshift/pikkanode/internal/validator"
)

type GetRequest struct {
	ID string `json:"id"`
}

func (req *GetRequest) Valid() error {
	v := validator.New()
	v.Must(req.ID != "", "id required")
	v.Must(govalidator.IsNumeric(req.ID), "invalid id")

	return v.Error()
}

type CommentItem struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
	User      struct {
		Username string           `json:"username"`
		Photo    file.DownloadURL `json:"photo"`
	} `json:"user"`
}

type GetResult struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	Detail    string           `json:"detail"`
	Photo     file.DownloadURL `json:"photo"`
	Tags      []string         `json:"tags"`
	Username  string           `json:"username"`
	Comments  []*CommentItem   `json:"comments"`
	CreatedAt time.Time        `json:"createdAt"`
}

func Get(ctx context.Context, req *GetRequest) (*GetResult, error) {
	var r GetResult

	{
		// language=SQL
		err := pgctx.QueryRow(ctx, `
			select
				w.id, w.name, w.detail, w.photo, w.tags, w.created_at,
				u.username
			from works w
				left join users u on w.user_id = u.id
			where w.id = $1
		`, req.ID).Scan(
			&r.ID, &r.Name, &r.Detail, &r.Photo, pq.Array(&r.Tags), &r.CreatedAt,
			&r.Username,
		)
		if err == sql.ErrNoRows {
			return nil, errWorkNotFound
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
			where c.work_id = $1
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

		rows.Close()
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
	v.Must(govalidator.IsNumeric(req.ID), "invalid id")

	return v.Error()
}

func Favorite(ctx context.Context, req *FavoriteRequest) (*struct{}, error) {
	userID := session.GetUserID(ctx)
	if userID == "" {
		return nil, errInvalidCredentials
	}

	if req.Favorite {
		// language=SQL
		_, err := pgctx.Exec(ctx, `
		insert into favorites
			(user_id, work_id)
		values
			($1, $2)
		on conflict do nothing
	`, userID, req.ID)
		if pgsql.IsForeignKeyViolation(err, "favorites_work_id_fkey") {
			return nil, errWorkNotFound
		}
		if err != nil {
			return nil, err
		}
	} else {
		// language=SQL
		_, err := pgctx.Exec(ctx, `
			delete from favorites where user_id = $1 and work_id = $2
		`, userID, req.ID)
		if err != nil {
			return nil, err
		}
	}

	return new(struct{}), nil
}

type CommentRequest struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

func (req *CommentRequest) Valid() error {
	v := validator.New()
	v.Must(req.ID != "", "id required")
	v.Must(govalidator.IsNumeric(req.ID), "invalid id")
	v.Must(req.Content != "", "content required")
	v.Must(utf8.RuneCountInString(req.Content) <= 255, "content length maximum 255 characters")

	return v.Error()
}

func PostComment(ctx context.Context, req *CommentRequest) (*struct{}, error) {
	userID := session.GetUserID(ctx)
	if userID == "" {
		return nil, errInvalidCredentials
	}

	// language=SQL
	_, err := pgctx.Exec(ctx, `
		insert into comments
			(user_id, work_id, content)
		values
			($1, $2, $3)
	`, userID, req.ID, req.Content)
	if pgsql.IsForeignKeyViolation(err, "comments_work_id_fkey") {
		return nil, errWorkNotFound
	}
	if err != nil {
		return nil, err
	}

	return new(struct{}), nil
}

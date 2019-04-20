package discovery

import (
	"context"
	"time"

	"github.com/acoshift/pgsql/pgctx"
	"github.com/lib/pq"

	"github.com/acoshift/pikkanode/internal/file"
	"github.com/acoshift/pikkanode/internal/paginate"
	"github.com/acoshift/pikkanode/internal/session"
)

type GetWorksRequest struct {
	Paginate paginate.Paginate `json:"paginate"`
}

type WorkItem struct {
	ID         string           `json:"id"`
	Name       string           `json:"name"`
	Detail     string           `json:"detail"`
	Photo      file.DownloadURL `json:"photo"`
	Tags       []string         `json:"tags"`
	CreatedAt  time.Time        `json:"createdAt"`
	IsFavorite bool             `json:"isFavorite"`
}

type GetWorksResult struct {
	List     []*WorkItem       `json:"list"`
	Paginate paginate.Paginate `json:"paginate"`
}

func GetWorks(ctx context.Context, req *GetWorksRequest) (*GetWorksResult, error) {
	userID := session.GetUserID(ctx)

	var r GetWorksResult
	{
		err := req.Paginate.CountFrom(func() (cnt int64, err error) {
			// language=SQL
			err = pgctx.QueryRow(ctx, `
				select count(*) from works
			`).Scan(&cnt)
			return
		})
		if err != nil {
			return nil, err
		}
	}

	{
		// language=SQL
		rows, err := pgctx.Query(ctx, `
			select
				w.id, w.name, w.detail, w.photo, w.tags, w.created_at,
				f.work_id is not null as is_favorite
			from works w
				left join favorites f on w.id = f.work_id and ($3 != '' and f.user_id = $3::uuid)
			order by w.created_at desc
			offset $1 limit $2
		`, req.Paginate.Offset(), req.Paginate.Limit(), userID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		r.List = make([]*WorkItem, 0)
		r.Paginate = req.Paginate

		for rows.Next() {
			var x WorkItem
			err := rows.Scan(
				&x.ID, &x.Name, &x.Detail, &x.Photo, pq.Array(&x.Tags), &x.CreatedAt,
				&x.IsFavorite,
			)
			if err != nil {
				return nil, err
			}
			r.List = append(r.List, &x)
		}

		if err := rows.Err(); err != nil {
			return nil, err
		}

		rows.Close()
	}

	return &r, nil
}

package discovery

import (
	"context"
	"time"

	"github.com/acoshift/pgsql/pgctx"
	"github.com/lib/pq"

	"github.com/acoshift/pikkanode/internal/file"
	"github.com/acoshift/pikkanode/internal/paginate"
)

type GetWorksRequest struct {
	Paginate paginate.Paginate `json:"paginate"`
}

type WorkItem struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	Detail    string           `json:"detail"`
	Photo     file.DownloadURL `json:"photo"`
	Tags      []string         `json:"tags"`
	CreatedAt time.Time        `json:"createdAt"`
}

type GetWorksResult struct {
	List     []*WorkItem       `json:"list"`
	Paginate paginate.Paginate `json:"paginate"`
}

func GetWorks(ctx context.Context, req *GetWorksRequest) (*GetWorksResult, error) {
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
				id, name, detail, photo, tags, created_at
			from works
			order by created_at desc
			offset $1 limit $2
		`, req.Paginate.Offset(), req.Paginate.Limit())
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

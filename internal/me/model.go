package me

import (
	"context"

	"github.com/acoshift/pgsql/pgctx"
	"github.com/lib/pq"
)

func setUserPhoto(ctx context.Context, userID string, photo string) error {
	// language=SQL
	_, err := pgctx.Exec(ctx, `
		update users
		set photo = $2
		where id = $1
	`, userID, photo)
	return err
}

type insertWorkPhotoParam struct {
	UserID string
	Name   string
	Detail string
	Photo  string
	Tags   []string
}

func insertWorkPhoto(ctx context.Context, x *insertWorkPhotoParam) (id int64, err error) {
	// language=SQL
	err = pgctx.QueryRow(ctx, `
		insert into works
			(user_id, name, detail, photo, tags)
		values
			($1, $2, $3, $4, $5)
		returning id
	`, x.UserID, x.Name, x.Detail, x.Photo, pq.Array(x.Tags)).Scan(&id)
	return
}

type updateWorkParam struct {
	ID     string
	Name   string
	Detail string
	Tags   []string
}

func updateWork(ctx context.Context, x *updateWorkParam) error {
	// language=SQL
	_, err := pgctx.Exec(ctx, `
		update works
		set
			name = $2,
			detail = $3,
			tags = $4
		where id = $1
	`, x.ID, x.Name, x.Detail, pq.Array(x.Tags))
	return err
}

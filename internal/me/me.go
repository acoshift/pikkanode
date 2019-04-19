package me

import (
	"bytes"
	"context"
	"database/sql"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/acoshift/arpc"
	"github.com/acoshift/pgsql/pgctx"
	"github.com/asaskevich/govalidator"
	"github.com/lib/pq"

	"github.com/acoshift/pikkanode/internal/file"
	"github.com/acoshift/pikkanode/internal/image"
	"github.com/acoshift/pikkanode/internal/paginate"
	"github.com/acoshift/pikkanode/internal/session"
	"github.com/acoshift/pikkanode/internal/validator"
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
	Username string           `json:"username"`
	Photo    file.DownloadURL `json:"photo"`
}

func Profile(ctx context.Context, _ *ProfileRequest) (*ProfileResult, error) {
	userID := session.GetUserID(ctx)
	if userID == "" {
		return nil, errInvalidCredentials
	}

	var r ProfileResult
	// language=SQL
	err := pgctx.QueryRow(ctx, `
		select username, photo
		from users
		where id = $1
	`, userID).Scan(
		&r.Username, &r.Photo,
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

type UploadProfilePhotoRequest struct {
	Photo *multipart.FileHeader
}

func (req *UploadProfilePhotoRequest) UnmarshalJSON(_ []byte) error {
	return arpc.ErrUnsupported
}

func (req *UploadProfilePhotoRequest) UnmarshalMultipartForm(v *multipart.Form) error {
	fp := v.File["photo"]
	if len(fp) == 1 {
		req.Photo = fp[0]
	}

	return nil
}

func (req *UploadProfilePhotoRequest) Valid() error {
	v := validator.New()
	v.Must(req.Photo != nil, "photo required")
	v.Must(image.Valid(req.Photo), "invalid photo")

	return v.Error()
}

func UploadProfilePhoto(ctx context.Context, req *UploadProfilePhotoRequest) (*struct{}, error) {
	userID := session.GetUserID(ctx)
	if userID == "" {
		return nil, errInvalidCredentials
	}

	ext := image.Ext(req.Photo)

	fp, err := req.Photo.Open()
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	var buf bytes.Buffer
	err = image.Profile(ctx, &buf, fp, ext)
	if err != nil {
		return nil, err
	}

	fn := file.GenerateFilename(ext)

	err = file.Store(ctx, file.File{
		Reader:      &buf,
		Name:        fn,
		ContentType: image.ContentType(ext),
	})
	if err != nil {
		return nil, err
	}

	err = setUserPhoto(ctx, userID, fn)
	if err != nil {
		return nil, err
	}

	return new(struct{}), nil
}

type RemoveWorkRequest struct {
	ID string `json:"id"`
}

func (req *RemoveWorkRequest) Valid() error {
	v := validator.New()
	v.Must(req.ID != "", "id required")
	v.Must(govalidator.IsNumeric(req.ID), "invalid id")

	return v.Error()
}

func RemoveWork(ctx context.Context, req *RemoveWorkRequest) (*struct{}, error) {
	userID := session.GetUserID(ctx)
	if userID == "" {
		return nil, errInvalidCredentials
	}

	// language=SQL
	_, err := pgctx.Exec(ctx, `
		delete from works where user_id = $1 and id = $2
	`, userID, req.ID)
	if err != nil {
		return nil, err
	}

	return new(struct{}), nil
}

type GetMyWorksRequest struct {
	Paginate paginate.Paginate `json:"paginate"`
}

type MyWorkItem struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	Detail    string           `json:"detail"`
	Photo     file.DownloadURL `json:"photo"`
	Tags      []string         `json:"tags"`
	CreatedAt time.Time        `json:"createdAt"`
}

type GetMyWorksResult struct {
	List     []*MyWorkItem     `json:"list"`
	Paginate paginate.Paginate `json:"paginate"`
}

func GetMyWorks(ctx context.Context, req *GetMyWorksRequest) (*GetMyWorksResult, error) {
	userID := session.GetUserID(ctx)
	if userID == "" {
		return nil, errInvalidCredentials
	}

	var r GetMyWorksResult
	{
		err := req.Paginate.CountFrom(func() (cnt int64, err error) {
			// language=SQL
			err = pgctx.QueryRow(ctx, `
				select count(*) from works where user_id = $1
			`, userID).Scan(&cnt)
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
			where user_id = $3
			offset $1 limit $2
		`, req.Paginate.Offset(), req.Paginate.Limit(), userID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		r.List = make([]*MyWorkItem, 0)
		r.Paginate = req.Paginate

		for rows.Next() {
			var x MyWorkItem
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

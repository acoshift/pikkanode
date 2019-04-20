package me

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"strconv"
	"time"
	"unicode/utf8"

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

type GetMyFavoriteWorksRequest struct {
	Paginate paginate.Paginate `json:"paginate"`
}

type GetMyFavoriteWorksResult struct {
	List     []*MyWorkItem     `json:"list"`
	Paginate paginate.Paginate `json:"paginate"`
}

func GetMyFavoriteWorks(ctx context.Context, req *GetMyFavoriteWorksRequest) (*GetMyFavoriteWorksResult, error) {
	userID := session.GetUserID(ctx)
	if userID == "" {
		return nil, errInvalidCredentials
	}

	var r GetMyFavoriteWorksResult
	{
		err := req.Paginate.CountFrom(func() (cnt int64, err error) {
			// language=SQL
			err = pgctx.QueryRow(ctx, `
				select count(*) from favorites where user_id = $1
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
				w.id, w.name, w.detail, w.photo, w.tags, w.created_at
			from favorites f
				left join works w on f.work_id = w.id
			where f.user_id = $3
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

type CreateWorkRequest struct {
	Name   string
	Detail string
	Photo  *multipart.FileHeader
	Tags   []string
}

func (req *CreateWorkRequest) UnmarshalJSON(_ []byte) error {
	return arpc.ErrUnsupported
}

func (req *CreateWorkRequest) UnmarshalMultipartForm(v *multipart.Form) error {
	if p := v.Value["name"]; len(p) == 1 {
		req.Name = p[0]
	}
	if p := v.Value["detail"]; len(p) == 1 {
		req.Detail = p[0]
	}
	req.Tags = v.Value["tags"]

	fp := v.File["photo"]
	if len(fp) == 1 {
		req.Photo = fp[0]
	}

	return nil
}

func (req *CreateWorkRequest) Valid() error {
	v := validator.New()
	v.Must(req.Name != "", "name required")
	v.Must(utf8.RuneCountInString(req.Name) <= 128, "name maximum 128 characters")
	v.Must(utf8.RuneCountInString(req.Detail) <= 1024, "name maximum 1024 characters")
	v.Must(req.Photo != nil, "photo required")
	v.Must(image.Valid(req.Photo), "invalid photo")
	for i, t := range req.Tags {
		v.Must(validator.IsTag(t), fmt.Sprintf("tags[%d] is not valid tag", i))
	}

	return v.Error()
}

type CreateWorkResult struct {
	ID     string           `json:"id"`
	Name   string           `json:"name"`
	Detail string           `json:"detail"`
	Photo  file.DownloadURL `json:"photo"`
	Tags   []string         `json:"tags"`
}

func CreateWork(ctx context.Context, req *CreateWorkRequest) (*CreateWorkResult, error) {
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

	req.Tags = append([]string{}, req.Tags...)
	id, err := insertWorkPhoto(ctx, &insertWorkPhotoParam{
		UserID: userID,
		Name:   req.Name,
		Detail: req.Detail,
		Photo:  fn,
		Tags:   req.Tags,
	})
	if err != nil {
		return nil, err
	}

	var r CreateWorkResult
	r.ID = strconv.FormatInt(id, 10)
	r.Name = req.Name
	r.Detail = req.Detail
	r.Photo = file.DownloadURL(fn)
	r.Tags = req.Tags
	return &r, nil
}

type UpdateWorkRequest struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Detail string   `json:"detail"`
	Tags   []string `json:"tags"`
}

func (req *UpdateWorkRequest) Valid() error {
	v := validator.New()
	v.Must(req.ID != "", "id required")
	v.Must(govalidator.IsNumeric(req.ID), "invalid id")
	v.Must(req.Name != "", "name required")
	v.Must(utf8.RuneCountInString(req.Name) <= 128, "name maximum 128 characters")
	v.Must(utf8.RuneCountInString(req.Detail) <= 1024, "name maximum 1024 characters")
	for i, t := range req.Tags {
		v.Must(validator.IsTag(t), fmt.Sprintf("tags[%d] is not valid tag", i))
	}

	return v.Error()
}

func UpdateWork(ctx context.Context, req *UpdateWorkRequest) (*struct{}, error) {
	userID := session.GetUserID(ctx)
	if userID == "" {
		return nil, errInvalidCredentials
	}

	{
		isExists := false
		// language=SQL
		err := pgctx.QueryRow(ctx, `
			select exists(
				select 1
				from works
				where user_id = $1 and id = $2
			)
		`, userID, req.ID).Scan(&isExists)
		if err != nil {
			return nil, err
		}

		if !isExists {
			return nil, errWorkNotFound
		}
	}

	{
		req.Tags = append([]string{}, req.Tags...)
		err := updateWork(ctx, &updateWorkParam{
			ID:     req.ID,
			Name:   req.Name,
			Detail: req.Detail,
			Tags:   req.Tags,
		})
		if err != nil {
			return nil, err
		}
	}

	return new(struct{}), nil
}

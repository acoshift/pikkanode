package me

import (
	"bytes"
	"context"
	"database/sql"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/acoshift/arpc"
	"github.com/acoshift/pgsql/pgctx"

	"github.com/acoshift/pikkanode/internal/file"
	"github.com/acoshift/pikkanode/internal/image"
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

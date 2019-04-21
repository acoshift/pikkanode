package image

import (
	"context"
	"io"
	"mime"
	"mime/multipart"

	"github.com/acoshift/arpc"
	"github.com/disintegration/imaging"
	"golang.org/x/sync/semaphore"
)

var (
	ErrInvalidType = arpc.NewError("invalid image type")
	ErrTooLarge    = arpc.NewError("image too large")
)

var contentTypeExt = map[string]string{
	"image/jpeg": "jpg",
	"image/jpg":  "jpg",
	"image/png":  "png",
	"image/gif":  "gif",
}

func Valid(fh *multipart.FileHeader) error {
	if fh == nil {
		return ErrInvalidType
	}

	// validate content type
	mt, _, _ := mime.ParseMediaType(fh.Header.Get("Content-Type"))
	if _, ok := contentTypeExt[mt]; !ok {
		return ErrInvalidType
	}

	if fh.Size == 0 {
		return ErrInvalidType
	}

	if fh.Size > 30<<20 { // 30 MiB
		return ErrTooLarge
	}

	return nil
}

func Ext(fh *multipart.FileHeader) string {
	mt, _, _ := mime.ParseMediaType(fh.Header.Get("Content-Type"))
	return contentTypeExt[mt]
}

func ContentType(ext string) string {
	for ct, t := range contentTypeExt {
		if t == ext {
			return ct
		}
	}
	return ""
}

var sem = semaphore.NewWeighted(10)

func Profile(ctx context.Context, w io.Writer, r io.Reader, ext string) error {
	ft, err := imaging.FormatFromExtension(ext)
	if err != nil {
		return ErrInvalidType
	}

	// only profile only jpeg or gif
	switch ft {
	case imaging.GIF:
	case imaging.JPEG:
	default:
		ft = imaging.JPEG
	}

	sem.Acquire(ctx, 1)
	defer sem.Release(1)

	img, err := imaging.Decode(r)
	if err != nil {
		return ErrInvalidType
	}

	img = imaging.Fill(img, 250, 250, imaging.Center, imaging.Lanczos)
	return imaging.Encode(w, img, ft)
}

func Sanitize(ctx context.Context, w io.Writer, r io.Reader, ext string) error {
	ft, err := imaging.FormatFromExtension(ext)
	if err != nil {
		return ErrInvalidType
	}

	sem.Acquire(ctx, 1)
	defer sem.Release(1)

	img, err := imaging.Decode(r)
	if err != nil {
		return err
	}

	return imaging.Encode(w, img, ft)
}

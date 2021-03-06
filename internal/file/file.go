package file

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/gofrs/uuid"

	"github.com/acoshift/pikkanode/internal/config"
)

var (
	client         = config.StorageClient()
	bucketName     = config.String("storage_bucket")
	bucketBasePath = config.String("storage_base")
	bucketHandle   = client.Bucket(bucketName)
	baseURL        = config.BaseURL()
)

const BasePath = "/u"

func GenerateFilename(ext string) string {
	ext = strings.TrimPrefix(ext, ".")
	return uuid.Must(uuid.NewV4()).String() + "." + ext
}

func Serve(ctx context.Context, w http.ResponseWriter, filename string) error {
	fn := path.Join(bucketBasePath, filename)

	obj := bucketHandle.Object(fn)
	rd, err := obj.NewReader(ctx)
	if err != nil {
		return err
	}
	defer rd.Close()

	w.Header().Set("Content-Type", rd.Attrs.ContentType)

	_, err = io.Copy(w, rd)
	return err
}

func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filename := r.URL.Path
		Serve(r.Context(), w, filename)
	})
}

type File struct {
	io.Reader
	Name        string
	ContentType string
}

func Store(ctx context.Context, f File) error {
	fn := path.Join(bucketBasePath, f.Name)
	obj := bucketHandle.Object(fn)
	w := obj.NewWriter(ctx)
	defer w.Close()

	w.ContentType = f.ContentType

	_, err := io.Copy(w, f)
	if err != nil {
		return err
	}

	return w.Close()
}

type DownloadURL string

func (s DownloadURL) MarshalJSON() ([]byte, error) {
	if s == "" {
		return json.Marshal("")
	}
	p := baseURL + path.Join(BasePath, string(s))
	return json.Marshal(p)
}

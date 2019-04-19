package picture

import (
	"net/http"

	"github.com/acoshift/arpc"
)

var (
	errPictureNotFound = arpc.NewError(http.StatusBadRequest, "picture not found")
)

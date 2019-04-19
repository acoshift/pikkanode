package picture

import (
	"net/http"

	"github.com/acoshift/arpc"
)

var (
	errInvalidCredentials = arpc.NewError(http.StatusUnauthorized, "invalid credentials")
	errPictureNotFound    = arpc.NewError(http.StatusBadRequest, "picture not found")
)

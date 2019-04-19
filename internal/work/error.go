package work

import (
	"net/http"

	"github.com/acoshift/arpc"
)

var (
	errInvalidCredentials = arpc.NewError(http.StatusUnauthorized, "invalid credentials")
	errWorkNotFound       = arpc.NewError(http.StatusBadRequest, "work not found")
)

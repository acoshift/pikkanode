package auth

import (
	"net/http"

	"github.com/acoshift/arpc"
)

var (
	errUsernameNotAvailable = arpc.NewError(http.StatusBadRequest, "username not available")
	errInvalidCredentials   = arpc.NewError(http.StatusUnauthorized, "invalid credentials")
)

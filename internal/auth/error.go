package auth

import (
	"github.com/acoshift/arpc"
)

var (
	errUsernameNotAvailable = arpc.NewError("username not available")
	errInvalidCredentials   = arpc.NewError("invalid credentials")
)

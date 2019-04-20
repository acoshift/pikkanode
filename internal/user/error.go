package user

import (
	"github.com/acoshift/arpc"
)

var (
	errInvalidCredentials = arpc.NewError("invalid credentials")
)

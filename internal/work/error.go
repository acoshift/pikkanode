package work

import (
	"github.com/acoshift/arpc"
)

var (
	errInvalidCredentials = arpc.NewError("invalid credentials")
	errWorkNotFound       = arpc.NewError("photo not found")
)

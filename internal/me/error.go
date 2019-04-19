package me

import (
	"net/http"

	"github.com/acoshift/arpc"
)

var (
	errInvalidCredentials = arpc.NewError(http.StatusUnauthorized, "invalid credentials")
)

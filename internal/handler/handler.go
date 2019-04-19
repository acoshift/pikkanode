package handler

import (
	"net/http"

	"github.com/acoshift/arpc"
	"github.com/acoshift/middleware"
	"github.com/acoshift/pgsql/pgctx"

	"github.com/acoshift/pikkanode/internal/auth"
	"github.com/acoshift/pikkanode/internal/config"
	"github.com/acoshift/pikkanode/internal/file"
	"github.com/acoshift/pikkanode/internal/me"
	"github.com/acoshift/pikkanode/internal/picture"
	"github.com/acoshift/pikkanode/internal/session"
	"github.com/acoshift/pikkanode/internal/user"
)

func New() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", arpc.NotFoundHandler())
	mux.Handle(file.BasePath+"/", http.StripPrefix(file.BasePath, file.Handler()))

	mux.Handle("/auth/signUp", arpc.Handler(auth.SignUp))
	mux.Handle("/auth/signIn", arpc.Handler(auth.SignIn))
	mux.Handle("/auth/signOut", arpc.Handler(auth.SignOut))

	mux.Handle("/me/profile", arpc.Handler(me.Profile))
	mux.Handle("/me/uploadProfilePhoto", arpc.Handler(me.UploadProfilePhoto))

	mux.Handle("/user/profile", arpc.Handler(user.Profile))
	mux.Handle("/user/follow", arpc.Handler(user.Follow))

	mux.Handle("/picture/get", arpc.Handler(picture.Get))
	mux.Handle("/picture/favorite", arpc.Handler(picture.Favorite))
	mux.Handle("/picture/postComment", arpc.Handler(picture.PostComment))

	return middleware.Chain(
		session.Middleware(),
		pgctx.Middleware(config.DB()),
	)(mux)
}

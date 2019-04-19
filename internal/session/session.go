package session

import (
	"context"
	"time"

	"github.com/acoshift/middleware"
	"github.com/moonrhythm/session"
	redisstore "github.com/moonrhythm/session/store/goredis"

	"github.com/acoshift/pikkanode/internal/config"
)

func Middleware() middleware.Middleware {
	return session.Middleware(session.Config{
		Secure:      session.PreferSecure,
		IdleTimeout: 7 * 24 * time.Hour,
		MaxAge:      30 * 24 * time.Hour,
		Rolling:     true,
		Proxy:       true,
		Path:        "/",
		HTTPOnly:    true,
		Store: redisstore.New(redisstore.Config{
			Prefix: config.RedisPrefix(),
			Client: config.RedisClient(),
		}),
	})
}

const sessName = "s"

func Get(ctx context.Context) *session.Session {
	s, err := session.Get(ctx, sessName)
	if err != nil {
		panic(err)
	}
	return s
}

func GetUserID(ctx context.Context) string {
	s := Get(ctx)
	return s.GetString("user_id")
}

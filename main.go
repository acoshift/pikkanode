package main

import (
	"net/http"
	"time"

	"github.com/moonrhythm/parapet"
	"github.com/moonrhythm/parapet/pkg/cors"
	"github.com/moonrhythm/parapet/pkg/healthz"
	"github.com/moonrhythm/parapet/pkg/host"
	"github.com/moonrhythm/parapet/pkg/location"
	"github.com/moonrhythm/parapet/pkg/redirect"

	"github.com/acoshift/pikkanode/internal/config"
	"github.com/acoshift/pikkanode/internal/handler"
)

func main() {
	h := handler.New()

	svc := parapet.NewBackend()
	if !config.Dev() {
		svc.Use(health())
		svc.Use(redirect.HTTPS())
		svc.Use(cors.CORS{
			MaxAge:           time.Hour,
			AllowCredentials: true,
			AllowHeaders:     []string{"Content-Type"},
			AllowMethods:     []string{"POST"},
			AllowOrigins: []string{
				"http://localhost:8080",
				"http://localhost:8000",
				"http://localhost:3000",
				"http://127.0.0.1:8080",
				"http://127.0.0.1:8000",
				"http://127.0.0.1:3000",
				"http://0.0.0.0:8080",
				"http://0.0.0.0:8000",
				"http://0.0.0.0:3000",
			},
		})
	}

	http.ListenAndServe(":8080", h)
}

func health() parapet.Middleware {
	h := host.NewCIDR("0.0.0.0/0")
	l := location.Exact("/healthz")
	l.Use(healthz.New())
	h.Use(l)
	return h
}

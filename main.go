package main

import (
	"log"
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
	svc := parapet.NewBackend()
	svc.Use(health())
	if !config.Dev() {
		svc.Use(redirect.HTTPS())
	}
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
	svc.Handler = handler.New()
	svc.Addr = ":8080"

	err := svc.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func health() parapet.Middleware {
	h := host.NewCIDR("0.0.0.0/0")
	l := location.Exact("/healthz")
	l.Use(healthz.New())
	h.Use(l)
	return h
}

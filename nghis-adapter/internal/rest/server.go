package rest

import (
	"net/http"
	"time"

	"github.com/arfis/waiting-room/nghis-adapter/internal/config/service"
	"github.com/arfis/waiting-room/nghis-adapter/internal/middleware"
	"github.com/arfis/waiting-room/nghis-adapter/internal/rest/register"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/dig"
)

func NewServer(configuration *service.Configuration, loggingMiddleware *middleware.LoggingMiddleware, diContainer *dig.Container) *http.Server {
	r := chi.NewRouter()
	r.Route(configuration.ServerContext, func(router chi.Router) {
		router.Use(middleware.Authorization)
		router.Use(chimiddleware.Compress(5))
		router.Use(middleware.RequestIdMiddleware)
		router.Use(loggingMiddleware.LoggingMiddleware)
		router.Use(chimiddleware.Timeout(configuration.HTTPTimeout))
		router.Use(middleware.Cors())
		register.Generated(router, diContainer)
	})
	http.Handle("/", r)

	return &http.Server{
		Addr:              ":" + configuration.ServerPort,
		Handler:           r,
		ReadHeaderTimeout: 2 * time.Second, // mitigate slow loris attack
	}
}

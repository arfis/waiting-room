package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/arfis/waiting-room/nghis-adapter/internal/client"
	"github.com/arfis/waiting-room/nghis-adapter/internal/config/service"
	nghisContext "github.com/arfis/waiting-room/nghis-adapter/internal/context"
	ngErrors "github.com/arfis/waiting-room/nghis-adapter/internal/errors"
	"github.com/arfis/waiting-room/nghis-adapter/internal/logger"
	"github.com/arfis/waiting-room/nghis-adapter/internal/middleware"
	"github.com/arfis/waiting-room/nghis-adapter/internal/rest"
	appointmentRest "github.com/arfis/waiting-room/nghis-adapter/internal/rest/handler/appointment"
	svcRest "github.com/arfis/waiting-room/nghis-adapter/internal/rest/handler/services"
	"github.com/arfis/waiting-room/nghis-adapter/internal/service/appointment"
	svcService "github.com/arfis/waiting-room/nghis-adapter/internal/service/services"
	"go.uber.org/dig"
)

type dependency struct {
	// Constructor is the provider of a component in the dependency container.
	Constructor any

	// BindToInterface is optional definition to which interface should be the Constructor bound.
	// Leave empty if no interface is used and no binding is required.
	BindToInterface any
}

// DIContainer collects all constructors below and its dependencies and
// creates a dependency container with autowired dependencies.
func DIContainer() *dig.Container {
	dependencies := []dependency{
		{Constructor: nghisContext.NewApplicationContext},
		{Constructor: service.NewConfiguration},
		{Constructor: logger.NewLogging},
		{Constructor: middleware.NewLoggingMiddleware},
		{Constructor: middleware.NewPagingMiddleware},
		{Constructor: ngErrors.NewResponseErrorHandler},

		{Constructor: client.NewHTTPClient},
		{Constructor: client.NewClinicalClient},
		{Constructor: client.NewPersonClient},

		{Constructor: rest.NewServer},

		{Constructor: appointmentRest.New},
		{Constructor: appointment.NewService},

		{Constructor: svcRest.New},
		{Constructor: svcService.NewService},
	}

	container := dig.New()
	for _, newDependency := range dependencies {
		if newDependency.BindToInterface != nil {
			if err := container.Provide(newDependency.Constructor, dig.As(newDependency.BindToInterface)); err != nil {
				panic(err)
			}
			continue
		}

		if err := container.Provide(newDependency.Constructor); err != nil {
			panic(err)
		}
	}

	if err := container.Provide(func() *dig.Container {
		return container
	}); err != nil {
		panic(err)
	}

	return container
}

func main() {
	diContainer := DIContainer()
	err := diContainer.Invoke(func(
		logger *slog.Logger,
		configuration *service.Configuration,
		cancelFunc context.CancelFunc,
		server *http.Server,
		httpClient *http.Client,
	) error {
		go func() {
			if err := server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
				logger.Error("finished", "service", configuration.ServiceName, "instance", configuration.ServiceInstanceName, "err", err)
			}
		}()

		logger.Info("running", "service", configuration.ServiceName, "instance", configuration.ServiceInstanceName,
			"port", configuration.ServerPort, "context", configuration.ServerContext)

		waitAndGracefullyStop(logger, make(chan os.Signal, 1), cancelFunc, server, httpClient)
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func waitAndGracefullyStop(logger *slog.Logger,
	signalChan chan os.Signal,
	cancelFunc context.CancelFunc,
	server *http.Server,
	httpClient *http.Client,
) {
	signal.Notify(signalChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
	logger.Info("received sigterm gracefully shutting down")

	cancelFunc()

	ctx, cancelHTTPServer := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelHTTPServer()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("http server forced to shutdown", "err", err)
	}

	httpClient.CloseIdleConnections()
}

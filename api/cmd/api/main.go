package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/dig"

	"github.com/arfis/waiting-room/internal/cardreader"
	"github.com/arfis/waiting-room/internal/config"
	ngErrors "github.com/arfis/waiting-room/internal/errors"
	"github.com/arfis/waiting-room/internal/middleware"
	queueService "github.com/arfis/waiting-room/internal/queue"
	"github.com/arfis/waiting-room/internal/repository"
	"github.com/arfis/waiting-room/internal/rest"
	configHandler "github.com/arfis/waiting-room/internal/rest/handler/configuration"
	kioskHandler "github.com/arfis/waiting-room/internal/rest/handler/kiosk"
	queueHandler "github.com/arfis/waiting-room/internal/rest/handler/queue"
	servicepointHandler "github.com/arfis/waiting-room/internal/rest/handler/servicepoint"
	configurationService "github.com/arfis/waiting-room/internal/service/configuration"
	kioskService "github.com/arfis/waiting-room/internal/service/kiosk"
	queueServiceGenerated "github.com/arfis/waiting-room/internal/service/queue"
	servicepointService "github.com/arfis/waiting-room/internal/service/servicepoint"
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
func DIContainer(cfg *config.Config) *dig.Container {
	dependencies := []dependency{
		// Configuration
		{Constructor: func() *config.Config {
			return cfg
		}},

		// Logger
		{Constructor: func() *slog.Logger {
			return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
		}},

		// Repository - try MongoDB first, fallback to mock
		{Constructor: func() repository.QueueRepository {
			// Try to connect to MongoDB using configuration
			repo, err := repository.NewMongoDBQueueRepository(cfg.GetMongoURI(), cfg.GetMongoDatabase())
			if err != nil {
				log.Printf("Failed to connect to MongoDB, using mock repository: %v", err)
				return repository.NewMockQueueRepository()
			}

			log.Println("Connected to MongoDB successfully")
			return repo
		}},

		// Core services
		{Constructor: func(repo repository.QueueRepository, cfg *config.Config, servicePointSvc *servicepointService.Service) *queueService.WaitingQueue {
			return queueService.NewWaitingQueue(repo, cfg, servicePointSvc)
		}},
		{Constructor: func(cfg *config.Config) *servicepointService.Service {
			return servicepointService.NewService(cfg)
		}},
		{Constructor: cardreader.NewService},

		// Middleware
		{Constructor: middleware.NewAuthorizationMiddleware},
		{Constructor: middleware.NewLoggingMiddleware},
		{Constructor: ngErrors.NewResponseErrorHandler},

		// Generated services (will be set up with broadcast function later)
		{Constructor: func(queueService *queueService.WaitingQueue) *kioskService.Service {
			return kioskService.New(queueService, nil)
		}},
		{Constructor: func(queueService *queueService.WaitingQueue) *queueServiceGenerated.Service {
			return queueServiceGenerated.New(queueService, nil)
		}},
		{Constructor: configurationService.New},

		// Generated handlers
		{Constructor: configHandler.New},
		{Constructor: kioskHandler.New},
		{Constructor: queueHandler.New},
		{Constructor: servicepointHandler.New},
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

	// Provide the container itself
	if err := container.Provide(func() *dig.Container {
		return container
	}); err != nil {
		panic(err)
	}

	return container
}

func main() {
	// Load configuration
	configPath := "config.yaml"
	if envConfigPath := os.Getenv("CONFIG_PATH"); envConfigPath != "" {
		configPath = envConfigPath
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Configuration loaded from: %s", configPath)
	log.Printf("Server will start on: %s", cfg.GetAddress())
	log.Printf("MongoDB URI: %s", cfg.GetMongoURI())
	log.Printf("WebSocket enabled: %v", cfg.WebSocket.Enabled)

	diContainer := DIContainer(cfg)

	// Create the server with the container and configuration
	server := rest.NewServer(diContainer, cfg)

	// Start the servicepoint cleanup routine
	diContainer.Invoke(func(servicePointSvc *servicepointService.Service) {
		ctx := context.Background()
		servicePointSvc.StartCleanupRoutine(ctx)
		log.Println("ServicePoint cleanup routine started")
	})

	go func() {
		log.Println("API listening on", server.Addr)
		if err := server.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Fatal("Server failed to start:", err)
			}
		}
	}()

	waitAndGracefullyStop(server)
}

func waitAndGracefullyStop(server *http.Server) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
	log.Println("received sigterm, gracefully shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
}

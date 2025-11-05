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

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/dig"

	"github.com/arfis/waiting-room/internal/cardreader"
	"github.com/arfis/waiting-room/internal/config"
	ngErrors "github.com/arfis/waiting-room/internal/errors"
	"github.com/arfis/waiting-room/internal/middleware"
	queueService "github.com/arfis/waiting-room/internal/queue"
	"github.com/arfis/waiting-room/internal/repository"
	"github.com/arfis/waiting-room/internal/rest"
	adminHandler "github.com/arfis/waiting-room/internal/rest/handler/admin"
	configHandler "github.com/arfis/waiting-room/internal/rest/handler/configuration"
	kioskHandler "github.com/arfis/waiting-room/internal/rest/handler/kiosk"
	queueHandler "github.com/arfis/waiting-room/internal/rest/handler/queue"
	servicepointHandler "github.com/arfis/waiting-room/internal/rest/handler/servicepoint"
	adminService "github.com/arfis/waiting-room/internal/service/admin"
	configService "github.com/arfis/waiting-room/internal/service/config"
	configurationService "github.com/arfis/waiting-room/internal/service/configuration"
	kioskService "github.com/arfis/waiting-room/internal/service/kiosk"
	queueServiceGenerated "github.com/arfis/waiting-room/internal/service/queue"
	servicepointService "github.com/arfis/waiting-room/internal/service/servicepoint"
	tenantService "github.com/arfis/waiting-room/internal/service/tenant"
	"github.com/arfis/waiting-room/internal/service/translation"
	webhookService "github.com/arfis/waiting-room/internal/service/webhook"
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
		{Constructor: func() repository.ConfigRepository {
			// Try to connect to MongoDB using configuration
			client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.GetMongoURI()))
			if err != nil {
				log.Printf("Failed to connect to MongoDB for config: %v", err)
				return nil
			}

			db := client.Database(cfg.GetMongoDatabase())
			repo := repository.NewMongoDBConfigRepository(db)
			log.Println("Connected to MongoDB for config successfully")
			return repo
		}},

		// Core services
		{Constructor: func(repo repository.QueueRepository, cfg *config.Config, servicePointSvc *servicepointService.Service, configService *configService.Service) *queueService.WaitingQueue {
			wq := queueService.NewWaitingQueue(repo, cfg, servicePointSvc)
			wq.SetConfigService(configService)
			return wq
		}},
		{Constructor: func(cfg *config.Config) *servicepointService.Service {
			return servicepointService.NewService(cfg)
		}},
		{Constructor: cardreader.NewService},

		// Middleware
		{Constructor: middleware.NewAuthorizationMiddleware},
		{Constructor: middleware.NewTenantMiddleware},
		{Constructor: middleware.NewLoggingMiddleware},
		{Constructor: ngErrors.NewResponseErrorHandler},

		// Translation service
		{Constructor: func(config *config.Config) *translation.DeepLTranslationService {
			return translation.NewDeepLTranslationService(config.DeepL)
		}},

		// Webhook service
		{Constructor: func(configService *configService.Service) *webhookService.Service {
			return webhookService.NewService(configService)
		}},

		// Generated services (will be set up with broadcast function later)
		{Constructor: func(queueService *queueService.WaitingQueue, config *config.Config, configService *configService.Service, webhookService *webhookService.Service, translationService *translation.DeepLTranslationService) *kioskService.Service {
			return kioskService.New(queueService, nil, config, configService, webhookService, translationService)
		}},
		{Constructor: func(queueService *queueService.WaitingQueue, webhookService *webhookService.Service) *queueServiceGenerated.Service {
			return queueServiceGenerated.New(queueService, nil, webhookService)
		}},
		{Constructor: func(cfg *config.Config, configService *configService.Service) *configurationService.Service {
			svc := configurationService.New(cfg)
			svc.SetConfigService(configService)
			return svc
		}},
		{Constructor: func(repo repository.ConfigRepository) *configService.Service {
			return configService.NewService(repo)
		}},
		{Constructor: func(repo repository.ConfigRepository) *tenantService.Service {
			return tenantService.NewService(repo)
		}},
		{Constructor: func(configService *configService.Service, translationService *translation.DeepLTranslationService, tenantService *tenantService.Service) *adminService.Service {
			return adminService.NewService(configService, translationService, tenantService)
		}},

		// Generated handlers
		{Constructor: adminHandler.New},
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

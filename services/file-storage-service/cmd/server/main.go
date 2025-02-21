package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	database "github.com/yaanno/upload-store-process/services/file-storage-service/internal/database/sqlite"
	repository "github.com/yaanno/upload-store-process/services/file-storage-service/internal/metadata"
	storageProvider "github.com/yaanno/upload-store-process/services/file-storage-service/internal/storage"
	handler "github.com/yaanno/upload-store-process/services/file-storage-service/internal/transport/http/handlers"
	router "github.com/yaanno/upload-store-process/services/file-storage-service/internal/transport/http/router"
	"github.com/yaanno/upload-store-process/services/file-storage-service/internal/upload"
	"github.com/yaanno/upload-store-process/services/shared/pkg/config"
	"github.com/yaanno/upload-store-process/services/shared/pkg/logger"
)

const serviceName = "file-storage-service"

func main() {

	// 1. Load Configuration
	cfg, err := loadConfiguration()
	if err != nil {
		fmt.Printf("Service failed to start due to configuration error: %v\n", err)
		os.Exit(1)
	}

	// 2. Initialize Logger
	log := logger.New(cfg.Logging)
	serviceLogger := log.WithService(serviceName)
	wrappedLogger := logger.Logger{Logger: serviceLogger}

	ctx := context.Background()

	// 3. Initialize Database
	db, err := database.InitializeTestDatabase(ctx)
	if err != nil {
		serviceLogger.Error().
			Err(err).
			Msg("Failed to initialize test database")
		os.Exit(1)
	}

	// 4. Initialize Storage Provider
	storage, err := initializeStorageProvider(cfg.Storage, wrappedLogger)
	if err != nil {
		serviceLogger.Error().Err(err).Msg("Failed to initialize storage provider, service exiting")
		os.Exit(1)
	}

	// 5. Initialize Repositories, Services, and Middleware
	metadataRepository, err := repository.NewRepository("sqlite", db, wrappedLogger)
	if err != nil {
		serviceLogger.Error().Err(err).Msg("Failed to initialize metadata repository, service exiting")
		os.Exit(1)
	}
	uploadService := upload.NewUploadService(metadataRepository, storage, wrappedLogger)
	// storageServiceServer := service.NewStorageService(wrappedLogger, storage)

	// 7. Initialize gRPC Server
	// grpcListener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port))
	// if err != nil {
	// 	serviceLogger.Error().Err(err).Str("host", cfg.Server.Host).Int("port", cfg.Server.Port).Msg("Failed to create gRPC listener, service exiting")
	// 	os.Exit(1)
	// }
	// grpcServer := grpc.NewServer(
	// 	grpc.UnaryInterceptor(interceptor.ValidationInterceptor()),
	// )
	// storagev1.RegisterFileStorageServiceServer(grpcServer, storageServiceServer)

	// 8. Initialize HTTP Server

	uploadHandler := handler.NewFileUploadHandler(wrappedLogger, uploadService)
	healthHandler := handler.NewHealthHandler(&serviceLogger)
	router := router.SetupRouter(uploadHandler, healthHandler)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.HttpServer.Host, cfg.HttpServer.Port),
		Handler: router,
	}

	// 9. Start Servers in Goroutines
	// go startGrpcServer(grpcServer, grpcListener, wrappedLogger, cfg.Server)
	go startHttpServer(httpServer, wrappedLogger, cfg.HttpServer)

	// 10. Graceful Shutdown Handling
	waitForShutdown(nil, httpServer, wrappedLogger)
}

func loadConfiguration() (*config.ServiceConfig, error) {
	defaults := &config.ServiceConfig{
		Logging: logger.LoggerConfig{
			Level:       "info",
			JSON:        true,
			Development: false,
		},
		Server: config.ServerConfig{
			Host: "0.0.0.0",
			Port: 50051,
		},
		HttpServer: config.HttpServerConfig{
			Host: "0.0.0.0",
			Port: 50052,
		},
		Database: config.DatabaseConfig{
			Driver: "sqlite",
			Path:   "/data/storage.db",
		},
		NATS: config.NATSConfig{
			Servers: []string{"nats://localhost:4222"},
			Cluster: "upload-store-cluster",
		},
		Storage: config.Storage{
			Provider:    "local",
			BasePath:    "./data/uploads",
			MaxFileSize: 10 * 1024 * 1024,
		},
		JWT: config.JWT{
			Secret: "secret_key",
			Issuer: "myservice",
		},
	}

	cfg, err := config.Load(serviceName, defaults)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}
	return cfg, nil
}

func validateConfig(cfg *config.ServiceConfig) error {
	if cfg.JWT.Secret == "" {
		return errors.New("JWT secret must be configured")
	}
	if cfg.Storage.BasePath == "" {
		return errors.New("storage base path must be configured")
	}
	return nil
}

func initializeStorageProvider(storageCfg config.Storage, serviceLogger logger.Logger) (storageProvider.Provider, error) {
	// if storageCfg.Provider == "local" {
	provider, err := storageProvider.NewProvider("local", storageCfg.BasePath, serviceLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize local storage provider: %w", err)
	}
	serviceLogger.Info().Str("provider", storageCfg.Provider).Str("basePath", storageCfg.BasePath).Msg("Storage provider initialized")
	return provider, nil
	// }

}

func startGrpcServer(grpcServer *grpc.Server, lis net.Listener, serviceLogger logger.Logger, serverCfg config.ServerConfig) {
	serviceLogger.Info().
		Str("host", serverCfg.Host).
		Int("port", serverCfg.Port).
		Msg("gRPC server starting")

	if err := grpcServer.Serve(lis); err != nil {
		serviceLogger.Error().Err(err).Msg("gRPC server failed")
		// Do NOT os.Exit here in goroutine. Let the main function handle shutdown.
		// Consider using channels to communicate errors back to main if needed for more complex error handling.
	}
}

func startHttpServer(httpServer *http.Server, serviceLogger logger.Logger, httpServerCfg config.HttpServerConfig) {
	serviceLogger.Info().
		Str("host", httpServerCfg.Host).
		Int("port", httpServerCfg.Port).
		Msg("HTTP server starting")

	if err := httpServer.ListenAndServe(); err != nil {
		serviceLogger.Info().Msg("HTTP server closed")
	}
}

func waitForShutdown(grpcServer *grpc.Server, httpServer *http.Server, serviceLogger logger.Logger) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	serviceLogger.Info().Msg("Shutting down servers...")
	grpcServer.GracefulStop()
	httpServer.Shutdown(context.Background())
	serviceLogger.Info().Msg("Server shutdown complete")
}

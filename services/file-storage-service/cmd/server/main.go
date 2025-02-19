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

	storagev1 "github.com/yaanno/upload-store-process/gen/go/filestorage/v1"
	interceptor "github.com/yaanno/upload-store-process/services/file-storage-service/interceptor"
	repository "github.com/yaanno/upload-store-process/services/file-storage-service/internal/repository/sqlite"
	"github.com/yaanno/upload-store-process/services/file-storage-service/internal/service"
	handler "github.com/yaanno/upload-store-process/services/file-storage-service/pkg/api/handler"
	router "github.com/yaanno/upload-store-process/services/file-storage-service/pkg/api/router"
	storageProvider "github.com/yaanno/upload-store-process/services/file-storage-service/pkg/filesystem"
	storageService "github.com/yaanno/upload-store-process/services/file-storage-service/pkg/objectstorage"
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

	// 3. Initialize Database
	testDatabase, err := repository.InitializeTestDatabase()
	if err != nil {
		serviceLogger.Error().
			Err(err).
			Msg("Failed to initialize test database")
		os.Exit(1)
	}

	db := testDatabase.GetDB()

	// 4. Initialize Storage Provider
	storage, err := initializeStorageProvider(cfg.Storage, wrappedLogger)
	if err != nil {
		serviceLogger.Error().Err(err).Msg("Failed to initialize storage provider, service exiting")
		os.Exit(1)
	}

	// 5. Initialize Repositories, Services, and Middleware
	fileMetadataRepository := repository.NewSQLiteFileMetadataRepository(db, wrappedLogger)
	fileUploadService := storageService.NewFileUploadServiceImpl(fileMetadataRepository, storage)
	fileStorageServiceServer := service.NewFileStorageService(fileMetadataRepository, wrappedLogger, storage)

	// 7. Initialize gRPC Server
	grpcLis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port))
	if err != nil {
		serviceLogger.Error().Err(err).Str("host", cfg.Server.Host).Int("port", cfg.Server.Port).Msg("Failed to create gRPC listener, service exiting")
		os.Exit(1)
	}
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.ValidationInterceptor()),
	)
	storagev1.RegisterFileStorageServiceServer(grpcServer, fileStorageServiceServer)

	// 8. Initialize HTTP Server

	uploadHandler := handler.NewFileUploadHandler(wrappedLogger, fileUploadService)
	healthHandler := handler.NewHealthHandler(&serviceLogger)
	router := router.SetupRouter(uploadHandler, healthHandler)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.HttpServer.Host, cfg.HttpServer.Port),
		Handler: router,
	}

	// 9. Start Servers in Goroutines
	go startGrpcServer(grpcServer, grpcLis, wrappedLogger, cfg.Server)
	go startHttpServer(httpServer, wrappedLogger, cfg.HttpServer)

	// 10. Graceful Shutdown Handling
	waitForShutdown(grpcServer, httpServer, wrappedLogger)
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

func initializeStorageProvider(storageCfg config.Storage, serviceLogger logger.Logger) (*storageProvider.LocalFilesystemStorage, error) {
	if storageCfg.Provider == "local" {
		provider := storageProvider.NewLocalFilesystemStorage(storageCfg.BasePath)
		serviceLogger.Info().Str("provider", storageCfg.Provider).Str("basePath", storageCfg.BasePath).Msg("Storage provider initialized")
		return provider, nil
	}
	// Add other storage providers here (e.g., S3, GCS) in the future.
	return nil, fmt.Errorf("unsupported storage provider: %s", storageCfg.Provider)
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

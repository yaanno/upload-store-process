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
	"github.com/yaanno/upload-store-process/services/file-storage-service/interceptor"
	database "github.com/yaanno/upload-store-process/services/file-storage-service/internal/database/sqlite"
	healthchecker "github.com/yaanno/upload-store-process/services/file-storage-service/internal/health"
	repository "github.com/yaanno/upload-store-process/services/file-storage-service/internal/metadata"
	storageProvider "github.com/yaanno/upload-store-process/services/file-storage-service/internal/storage"
	grpcHandler "github.com/yaanno/upload-store-process/services/file-storage-service/internal/transport/grpc/handlers"
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

	// 5. Initialize Repositories, Services, and Middleware
	metadataRepository, err := repository.NewRepository("sqlite", db, &wrappedLogger)
	if err != nil {
		serviceLogger.Error().Err(err).Msg("Failed to initialize metadata repository, service exiting")
		os.Exit(1)
	}
	metadataService := repository.NewMetadataService(metadataRepository, &wrappedLogger)

	// 4. Initialize Storage Provider
	storage, err := initializeStorageProvider(cfg.Storage, metadataService, &wrappedLogger)
	if err != nil {
		serviceLogger.Error().Err(err).Msg("Failed to initialize storage provider, service exiting")
		os.Exit(1)
	}

	uploadService := upload.NewUploadService(metadataRepository, storage, &wrappedLogger)

	healthChecker := healthchecker.NewHealthChecker(db, cfg.Storage.BasePath)

	// TODO: this should be the storageServiceServer because the handlers implement the same interface
	fileOperationHandler := grpcHandler.NewFileOperationdHandler(metadataService, &wrappedLogger)

	// 7. Initialize gRPC Server
	grpcServer, grpcListener, err := initializeGRPCServer(cfg, &wrappedLogger)
	if err != nil {
		serviceLogger.Error().Err(err).Msg("Failed to initialize gRPC server")
		os.Exit(1)
	}
	storagev1.RegisterFileStorageServiceServer(grpcServer, fileOperationHandler)

	// 8. Initialize HTTP Server

	uploadHandler := handler.NewFileUploadHandler(&wrappedLogger, uploadService)
	healthHandler := handler.NewHealthHandler(&serviceLogger, healthChecker)
	houseKeepingHandler := handler.NewHouseKeepingHandler(metadataService, &wrappedLogger)
	router := router.SetupRouter(uploadHandler, healthHandler, houseKeepingHandler)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.HttpServer.Host, cfg.HttpServer.Port),
		Handler: router,
	}

	// 9. Start Servers in Goroutines
	go startGrpcServer(grpcServer, grpcListener, &wrappedLogger, cfg.Server)
	go startHttpServer(httpServer, &wrappedLogger, cfg.HttpServer)

	// 10. Graceful Shutdown Handling
	waitForShutdown(grpcServer, httpServer, &wrappedLogger)
}

func loadConfiguration() (*config.ServiceConfig, error) {
	defaults := &config.ServiceConfig{
		Logging: logger.LoggerConfig{
			Level:       "info",
			JSON:        true,
			Development: true,
		},
		Server: config.ServerConfig{
			Host: "0.0.0.0",
			Port: 8001,
		},
		HttpServer: config.HttpServerConfig{
			Host: "0.0.0.0",
			Port: 8000,
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

func initializeStorageProvider(storageCfg config.Storage, metadataService repository.MetadataService, logger *logger.Logger) (storageProvider.Provider, error) {
	// if storageCfg.Provider == "local" {
	providerConfig := &storageProvider.Config{
		BasePath:        storageCfg.BasePath,
		MetadataService: metadataService,
	}
	provider, err := storageProvider.NewProvider("local", providerConfig, metadataService, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize local storage provider: %w", err)
	}
	logger.Info().Str("provider", storageCfg.Provider).Str("basePath", storageCfg.BasePath).Msg("Storage provider initialized")
	return provider, nil
	// }

}

func initializeGRPCServer(cfg *config.ServiceConfig, logger *logger.Logger) (*grpc.Server, net.Listener, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create gRPC listener: %w", err)
	}

	// Add more interceptors as needed
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(interceptor.ValidationInterceptor()),
		grpc.ChainUnaryInterceptor(
			interceptor.LoggingInterceptor(logger),
			interceptor.RecoveryInterceptor(),
		),
	}

	server := grpc.NewServer(opts...)

	return server, listener, nil
}

func startGrpcServer(grpcServer *grpc.Server, lis net.Listener, logger *logger.Logger, serverCfg config.ServerConfig) {
	logger.Info().
		Str("host", serverCfg.Host).
		Int("port", serverCfg.Port).
		Msg("gRPC server starting")

	if err := grpcServer.Serve(lis); err != nil {
		logger.Error().Err(err).Msg("gRPC server failed")
		// Do NOT os.Exit here in goroutine. Let the main function handle shutdown.
		// Consider using channels to communicate errors back to main if needed for more complex error handling.
	}
}

func startHttpServer(httpServer *http.Server, logger *logger.Logger, httpServerCfg config.HttpServerConfig) {
	logger.Info().
		Str("host", httpServerCfg.Host).
		Int("port", httpServerCfg.Port).
		Msg("HTTP server starting")

	if err := httpServer.ListenAndServe(); err != nil {
		logger.Info().Msg("HTTP server closed")
	}
}

func waitForShutdown(grpcServer *grpc.Server, httpServer *http.Server, logger *logger.Logger) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("Shutting down servers...")
	grpcServer.GracefulStop()
	logger.Info().Msg("gRPC server shutdown complete")
	httpServer.Shutdown(context.Background())
	logger.Info().Msg("Server shutdown complete")
}

package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	storagev1 "github.com/yaanno/upload-store-process/gen/go/filestorage/v1"
	"github.com/yaanno/upload-store-process/services/file-storage-service/internal/repository"
	"github.com/yaanno/upload-store-process/services/file-storage-service/internal/service"
	storageProvider "github.com/yaanno/upload-store-process/services/file-storage-service/internal/storage"
	"github.com/yaanno/upload-store-process/services/shared/pkg/auth"
	"github.com/yaanno/upload-store-process/services/shared/pkg/config"
	"github.com/yaanno/upload-store-process/services/shared/pkg/logger"
)

const serviceName = "file-storage-service"

func main() {
	// Default configuration
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
		Database: config.DatabaseConfig{
			Driver: "sqlite",
			Path:   "/data/storage.db",
		},
		NATS: config.NATSConfig{
			Servers: []string{"nats://localhost:4222"},
			Cluster: "upload-store-cluster",
		},
		Storage: config.Storage{
			Provider: "local",
			BasePath: "/data/uploads",
		},
		JWT: config.JWT{
			Secret: "secret_key",
			Issuer: "myservice",
		},
	}

	// Load configuration
	cfg, err := config.Load(serviceName, defaults)
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logger.New(cfg.Logging)
	serviceLogger := log.WithService(serviceName)

	// Create network listener
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d",
		cfg.Server.Host,
		cfg.Server.Port,
	))
	if err != nil {
		serviceLogger.Error().
			Err(err).
			Str("host", cfg.Server.Host).
			Int("port", cfg.Server.Port).
			Msg("Failed to create network listener")
		os.Exit(1)
	}

	wrappedLogger := logger.Logger{Logger: serviceLogger}

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// TODO: Replace with actual DB initialization
	testDatabase, err := repository.InitializeTestDatabase()
	if err != nil {
		serviceLogger.Error().
			Err(err).
			Msg("Failed to initialize test database")
		os.Exit(1)
	}

	db := testDatabase.GetDB()
	tokenGenerator := auth.NewTokenGenerator(cfg.JWT.Secret, cfg.JWT.Issuer)
	localFilesystemStorage := storageProvider.NewLocalFilesystemStorage(cfg.Storage.BasePath)
	fileMetadataRepository := repository.NewSQLiteFileMetadataRepository(db, wrappedLogger)
	fileStorageServiceServer := service.NewFileStorageService(fileMetadataRepository, wrappedLogger, localFilesystemStorage, tokenGenerator)
	storagev1.RegisterFileStorageServiceServer(grpcServer, fileStorageServiceServer)

	// Graceful shutdown setup
	go func() {
		serviceLogger.Info().
			Str("host", cfg.Server.Host).
			Int("port", cfg.Server.Port).
			Msg("File Storage Service starting")

		if err := grpcServer.Serve(lis); err != nil {
			serviceLogger.Error().
				Err(err).
				Msg("gRPC server failed")
			os.Exit(1)
		}
	}()

	// Shutdown handling
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	serviceLogger.Info().Msg("Shutting down server...")
	grpcServer.GracefulStop()
	serviceLogger.Info().Msg("Server shutdown complete")
}

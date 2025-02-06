package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

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

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// TODO: Register service implementations
	// fileStorageServer := service.NewFileStorageServer(cfg)
	// storagev1.RegisterFileStorageServiceServer(grpcServer, fileStorageServer)

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

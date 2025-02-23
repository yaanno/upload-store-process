package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	storagev1 "github.com/yaanno/upload-store-process/gen/go/filestorage/v1"
	"github.com/yaanno/upload-store-process/services/api-gateway-service/internal/handler"
	"github.com/yaanno/upload-store-process/services/api-gateway-service/internal/router"
	"github.com/yaanno/upload-store-process/services/shared/pkg/config"
	"github.com/yaanno/upload-store-process/services/shared/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const serviceName = "api-gateway-service"

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

	grpcServerPath := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	grpcClientConn, err := grpc.NewClient(
		grpcServerPath,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		serviceLogger.Error().Err(err).Msg("Failed to initialize gRPC client")
		os.Exit(1)
	}
	defer grpcClientConn.Close()

	service := storagev1.NewFileStorageServiceClient(grpcClientConn)

	uploadHandler := handler.NewFileUploadHandler(wrappedLogger, service)

	// 7. Initialize Health Check Handler
	healthCheckHandler := handler.NewHealthHandler(&wrappedLogger)

	router := router.SetupRouter(uploadHandler, healthCheckHandler)

	// 8. Initialize HTTP Server

	httpServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.HttpServer.Host, cfg.HttpServer.Port),
		Handler: router,
	}

	go startHttpServer(httpServer, wrappedLogger, cfg.HttpServer)

	waitForShutdown(httpServer, wrappedLogger)
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
			Port: 8080,
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
	return nil
}

func waitForShutdown(httpServer *http.Server, serviceLogger logger.Logger) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	serviceLogger.Info().Msg("Shutting down servers...")
	httpServer.Shutdown(context.Background())
	serviceLogger.Info().Msg("Server shutdown complete")
}

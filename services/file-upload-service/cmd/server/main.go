package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	uploadv1 "github.com/yaanno/upload-store-process/gen/go/fileupload/v1"
	"github.com/yaanno/upload-store-process/services/file-upload-service/internal/handler"
	"github.com/yaanno/upload-store-process/services/file-upload-service/internal/middleware"
	"github.com/yaanno/upload-store-process/services/shared/pkg/auth"
	"github.com/yaanno/upload-store-process/services/shared/pkg/config"
	"github.com/yaanno/upload-store-process/services/shared/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const serviceName = "file-upload-service"

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

	// Signal handling for graceful stopping
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	grpcClientConn, err := grpc.NewClient(cfg.Upload.GRPCAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		serviceLogger.Error().Err(err).Msg("Failed to initialize gRPC client")
		os.Exit(1)
	}
	defer grpcClientConn.Close()

	tokenGenerator := auth.NewTokenGenerator(cfg.JWT.Secret, cfg.JWT.Issuer)
	service := uploadv1.NewFileUploadServiceClient(grpcClientConn)
	jwtAuthMiddleware := middleware.NewJWTAuthMiddleware(wrappedLogger, cfg.Upload.MaxFileSize, tokenGenerator)
	uploadHandler := handler.NewFileUploadHandler(wrappedLogger, cfg.Upload.MaxFileSize, service)

	// 8. Initialize HTTP Server
	httpMux := http.NewServeMux()
	// httpMux.HandleFunc("/v1/files/upload", jwtAuthMiddleware.JWTAuthMiddleware(uploadHandler.HandleFileUpload))
	httpMux.HandleFunc("/v1/files/prepareupload", jwtAuthMiddleware.JWTAuthMiddleware(uploadHandler.PrepareUpload))
	httpMux.HandleFunc("/v1/files/getmetadata", jwtAuthMiddleware.JWTAuthMiddleware(uploadHandler.GetFileMetadata))
	httpMux.HandleFunc("/healthz", healthCheckHandler) // Health check endpoint

	httpServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.HttpServer.Host, cfg.HttpServer.Port),
		Handler: httpMux,
	}

	go startHttpServer(httpServer, wrappedLogger, cfg.HttpServer)

	// Wait for stop signal
	<-stop
	wrappedLogger.Info().Msg("Shutting down gracefully...")

	// Shutdown timeout
	shutdownTimeout := 10 * time.Second

	// Create a context with timeout for graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	// Shutdown HTTP server
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		wrappedLogger.Error().Err(err).Msg("HTTP server shutdown error")
	}

}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
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
			BasePath:    "/data/uploads",
			MaxFileSize: 10 * 1024 * 1024,
		},
		JWT: config.JWT{
			Secret: "secret_key",
			Issuer: "myservice",
		},
		Upload: config.Upload{
			MaxFileSize: 10 * 1024 * 1024,
			GRPCAddress: "localhost:50051",
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

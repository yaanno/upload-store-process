package router

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	handler "github.com/yaanno/upload-store-process/services/api-gateway-service/internal/handler"
)

func SetupRouter(uploadHandler handler.FileUploadHandler, healthCheckHandler handler.HealthHandler) chi.Router {
	r := chi.NewRouter()

	r.Use(httprate.LimitByIP(100, 1*time.Minute))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Route("/api/v1", func(r chi.Router) {
		// Health check
		r.Get("/healthz", healthCheckHandler.Healtz)
		// File operations
		r.Get("/files", uploadHandler.ListFiles)
		r.Get("/status/{id}", uploadHandler.GetFileStatus)
		r.Get("/metadata/{id}", uploadHandler.GetFileMetadata)
		r.Post("/prepare-upload", uploadHandler.PrepareUpload)
		r.Delete("/delete/{id}", uploadHandler.DeleteFile)
	})

	return r

}

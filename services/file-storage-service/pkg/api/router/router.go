package router

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/yaanno/upload-store-process/services/file-storage-service/pkg/api/handler"
)

func SetupRouter(uploadHandler handler.FileUploadHandler, healthCheckHandler handler.HealthHandler) chi.Router {
	r := chi.NewRouter()

	r.Use(httprate.LimitByIP(100, 1*time.Minute))
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
		// Bucket operations
		// r.Put("/bucket", uploadHandler.CreateBucket)
		// r.Delete("/bucket", uploadHandler.DeleteBucket)
		// r.Get("/buckets", uploadHandler.GetBuckets)
		// r.Get("/bucket", uploadHandler.GetBucket)
		// File operations
		r.Put("/upload", uploadHandler.CreateFile)
		r.Get("/get/:fileId", uploadHandler.GetFile)
		r.Delete("/delete/:fileId", uploadHandler.DeleteFile)
	})

	return r

}

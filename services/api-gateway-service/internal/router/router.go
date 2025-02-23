package router

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	handler "github.com/yaanno/upload-store-process/services/api-gateway-service/internal/handler"
)

func SetupRouter(uploadHandler handler.FileUploadHandler, healthCheckHandler handler.HealthHandler) chi.Router {
	r := chi.NewRouter()

	r.Route("/api/v1", func(r chi.Router) {
		// Health check
		r.Get("/healthz", healthCheckHandler.Healtz)
		// File operations
		r.Get("/files", uploadHandler.ListFiles)
		r.Get("/status/{id}", uploadHandler.GetFileStatus)
		r.Get("/metadata/{id}", uploadHandler.GetFileMetadata)
		r.Post("/prepare-upload", uploadHandler.PrepareUpload)
		r.Delete("/delete/{id}", uploadHandler.DeleteFile)
		// Other
		r.NotFound(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"msg": "error",
			})
		})
	})

	return r

}

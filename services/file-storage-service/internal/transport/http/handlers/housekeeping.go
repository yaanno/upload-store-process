package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/yaanno/upload-store-process/services/file-storage-service/internal/metadata"
	"github.com/yaanno/upload-store-process/services/shared/pkg/logger"
)

type HouseKeepingHandler struct {
	metadataService metadata.MetadataService
	logger          *logger.Logger
}

func NewHouseKeepingHandler(metadataService metadata.MetadataService, logger *logger.Logger) HouseKeepingHandler {
	return HouseKeepingHandler{
		metadataService: metadataService,
		logger:          logger,
	}
}

func (h *HouseKeepingHandler) CleanupMetadata(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	count, err := h.metadataService.CleanupExpiredMetadata(ctx)
	if err != nil {
		http.Error(w, "Cleanup failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"removed_count": count,
		"timestamp":     time.Now(),
	})
}

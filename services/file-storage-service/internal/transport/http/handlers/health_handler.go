package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"github.com/yaanno/upload-store-process/services/file-storage-service/internal/health"
)

type HealthHandler struct {
	logger  zerolog.Logger
	checker *health.HealthChecker
}

type HealthStatus struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

var ErrorStatus = "error"

func NewHealthHandler(logger *zerolog.Logger, checker *health.HealthChecker) HealthHandler {
	return HealthHandler{
		logger:  logger.With().Str("component", "health_handler").Logger(),
		checker: checker,
	}
}

func (h *HealthHandler) Healtz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	checks := h.checker.CheckHealth(ctx)

	status := http.StatusOK
	for _, check := range checks {
		if check.Status == health.StatusDown {
			h.logger.Warn().Str("component", check.Component).Msg("component is not ok")
			status = http.StatusServiceUnavailable
			break
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(checks); err != nil {
		h.logger.Error().Err(err).Msg("failed to encode response")
		http.Error(w, "failed to encode response", http.StatusServiceUnavailable)
	}
}

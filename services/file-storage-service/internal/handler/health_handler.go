package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type HealthHandler struct {
	logger zerolog.Logger
}

type HealthStatus struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

var ErrorStatus = "error"

func NewHealthHandler(logger *zerolog.Logger) HealthHandler {
	return HealthHandler{
		logger: logger.With().Str("component", "health_handler").Logger(),
	}
}

func (h *HealthHandler) Healtz(w http.ResponseWriter, r *http.Request) {
	_, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	status := HealthStatus{
		Status:    "ok",
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	if status.Status != "ok" {
		h.logger.Error().Msg("service is not ok")
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		h.logger.Info().Msg("service is ok")
		w.WriteHeader(http.StatusOK)
	}

	if err := json.NewEncoder(w).Encode(status); err != nil {
		h.logger.Error().Err(err).Msg("failed to encode response")
		http.Error(w, "failed to encode response", http.StatusServiceUnavailable)
	}
}

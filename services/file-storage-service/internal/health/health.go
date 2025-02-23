package health

import (
	"context"
	"database/sql"
	"os"
	"sync"
	"time"
)

type Status string

const (
	StatusUp   Status = "UP"
	StatusDown Status = "DOWN"
)

type HealthCheck struct {
	Component string
	Status    Status
	Error     string
	LastCheck time.Time
}

type HealthChecker struct {
	db          *sql.DB
	storagePath string
	mu          sync.RWMutex
	checks      map[string]*HealthCheck
}

func NewHealthChecker(db *sql.DB, storagePath string) *HealthChecker {
	return &HealthChecker{
		db:          db,
		storagePath: storagePath,
		checks:      make(map[string]*HealthCheck),
	}
}

func (h *HealthChecker) CheckHealth(ctx context.Context) map[string]*HealthCheck {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Check database
	h.checks["database"] = h.checkDatabase(ctx)

	// Check storage
	h.checks["storage"] = h.checkStorage()

	return h.checks
}

func (h *HealthChecker) checkDatabase(ctx context.Context) *HealthCheck {
	check := &HealthCheck{
		Component: "database",
		LastCheck: time.Now(),
	}

	err := h.db.PingContext(ctx)
	if err != nil {
		check.Status = StatusDown
		check.Error = err.Error()
		return check
	}

	check.Status = StatusUp
	return check
}

func (h *HealthChecker) checkStorage() *HealthCheck {
	check := &HealthCheck{
		Component: "storage",
		LastCheck: time.Now(),
	}

	_, err := os.Stat(h.storagePath)
	if err != nil {
		check.Status = StatusDown
		check.Error = err.Error()
		return check
	}

	// Check if directory is writable
	testFile := h.storagePath + "/.test"
	f, err := os.Create(testFile)
	if err != nil {
		check.Status = StatusDown
		check.Error = "storage not writable: " + err.Error()
		return check
	}
	f.Close()
	os.Remove(testFile)

	check.Status = StatusUp
	return check
}

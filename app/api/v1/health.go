// Package v1 is the first version of the interface
package v1

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/ray-d-song/merc/app/infra/httpx"
)

// HealthResponse describes the payload returned by the health endpoint.
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// RegisterHealthRoutes wires the /api/v1/health endpoint into the provided router group.
func RegisterHealthRoutes(router chi.Router) {
	router.Get("/health", healthHandler)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC(),
	})
}

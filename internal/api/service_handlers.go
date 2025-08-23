package api

import (
	"encoding/json"
	"net/http"

	"github.com/justanamir/medappoint/internal/db/gen"
)

type ServiceDeps struct {
	Q *gen.Queries
}

// ListServicesHandler handles GET /v1/services
func (d ServiceDeps) ListServicesHandler(w http.ResponseWriter, r *http.Request) {
	services, err := d.Q.ListServices(r.Context())
	if err != nil {
		println("ListServices error:", err.Error()) // TEMP debug
		http.Error(w, "failed to list services", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(services)
}

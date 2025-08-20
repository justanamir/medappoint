package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/justanamir/medappoint/internal/db/gen"
)

type AvailabilityDeps struct {
	Q *gen.Queries
}

// ListByProviderHandler handles GET /v1/availabilities?provider_id=123
func (d AvailabilityDeps) ListByProviderHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("provider_id")
	if q == "" {
		http.Error(w, "provider_id is required", http.StatusBadRequest)
		return
	}
	pid, err := strconv.ParseInt(q, 10, 64)
	if err != nil || pid <= 0 {
		http.Error(w, "provider_id must be a positive integer", http.StatusBadRequest)
		return
	}

	rows, err := d.Q.ListAvailabilitiesByProvider(r.Context(), pid)
	if err != nil {
		http.Error(w, "failed to list availabilities", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(rows)
}

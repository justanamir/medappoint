package api

import (
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
		ErrorJSON(w, http.StatusBadRequest, "provider_id is required", nil)
		return
	}
	pid, err := strconv.ParseInt(q, 10, 64)
	if err != nil || pid <= 0 {
		ErrorJSON(w, http.StatusBadRequest, "invalid provider_id", nil)
		return
	}

	rows, err := d.Q.ListAvailabilitiesByProvider(r.Context(), pid)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "failed to list availabilities", nil)
		return
	}

	JSON(w, http.StatusOK, rows)
}

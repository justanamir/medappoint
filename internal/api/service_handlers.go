package api

import (
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
		ErrorJSON(w, http.StatusInternalServerError, "failed to list services", nil)
		return
	}
	JSON(w, http.StatusOK, services)
}

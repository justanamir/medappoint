package api

import (
	"net/http"

	"github.com/justanamir/medappoint/internal/db/gen"
)

type ProviderDeps struct {
	Q *gen.Queries
}

// ListProvidersHandler handles GET /v1/providers
func (d ProviderDeps) ListProvidersHandler(w http.ResponseWriter, r *http.Request) {
	providers, err := d.Q.ListProviders(r.Context())
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "failed to list providers", nil)
		return
	}
	JSON(w, http.StatusOK, providers)
}

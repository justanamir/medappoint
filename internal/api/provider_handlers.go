package api

import (
	"encoding/json"
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
		println("ListProviders error:", err.Error()) // TEMP debug
		http.Error(w, "failed to list providers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(providers)
}

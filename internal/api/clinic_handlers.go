package api

import (
	"encoding/json"
	"net/http"

	"github.com/justanamir/medappoint/internal/db/gen"
)

type ClinicDeps struct {
	Q *gen.Queries
}

func (d ClinicDeps) ListClinicsHandler(w http.ResponseWriter, r *http.Request) {
	clinics, err := d.Q.ListClinics(r.Context())
	if err != nil {
		http.Error(w, "failed to list clinics", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(clinics)
}

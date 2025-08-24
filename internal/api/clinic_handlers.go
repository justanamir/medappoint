package api

import (
	"net/http"

	"github.com/justanamir/medappoint/internal/db/gen"
)

type ClinicDeps struct {
	Q *gen.Queries
}

func (d ClinicDeps) ListClinicsHandler(w http.ResponseWriter, r *http.Request) {
	clinics, err := d.Q.ListClinics(r.Context())
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "failed to list clinics", nil)
		return
	}
	JSON(w, http.StatusOK, clinics)
}

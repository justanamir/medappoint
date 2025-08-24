package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/justanamir/medappoint/internal/config"
	"github.com/justanamir/medappoint/internal/db/gen"
)

type MeDeps struct {
	Cfg config.Config
	Q   *gen.Queries
}

// GET /v1/me/appointments?limit=20&offset=0
func (d MeDeps) ListMyAppointments(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromCtx(r)
	if !ok || uid <= 0 {
		ErrorJSON(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	// map user -> patient
	patient, err := d.Q.GetPatientByUserID(r.Context(), uid)
	if err != nil {
		ErrorJSON(w, http.StatusNotFound, "patient profile not found", nil)
		return
	}

	// pagination
	q := r.URL.Query()
	limit := int32(20)
	offset := int32(0)
	if s := q.Get("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 100 {
			limit = int32(n)
		}
	}
	if s := q.Get("offset"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n >= 0 {
			offset = int32(n)
		}
	}

	loc, _ := time.LoadLocation("Asia/Kuala_Lumpur")
	now := time.Now().In(loc)

	rows, err := d.Q.ListUpcomingAppointmentsByPatient(r.Context(), gen.ListUpcomingAppointmentsByPatientParams{
		PatientID: patient.ID,
		StartTime: now,
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "failed to list appointments", nil)
		return
	}

	JSON(w, http.StatusOK, rows)
}

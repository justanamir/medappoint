package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/justanamir/medappoint/internal/config"
	"github.com/justanamir/medappoint/internal/db/gen"
)

type ProviderScheduleDeps struct {
	Cfg config.Config
	Q   *gen.Queries
}

// GET /v1/providers/{id}/appointments?date=YYYY-MM-DD
func (d ProviderScheduleDeps) ListProviderDayAppointments(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromCtx(r)
	if !ok || uid <= 0 {
		ErrorJSON(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	role, _ := RoleFromCtx(r)

	idStr := chi.URLParam(r, "id")
	providerID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || providerID <= 0 {
		ErrorJSON(w, http.StatusBadRequest, "invalid provider id", nil)
		return
	}

	if role == "provider" {
		myProv, err := d.Q.GetProviderByUserID(r.Context(), uid)
		if err != nil || myProv.ID != providerID {
			ErrorJSON(w, http.StatusForbidden, "forbidden", nil)
			return
		}
	}

	loc, _ := time.LoadLocation("Asia/Kuala_Lumpur")
	dateStr := r.URL.Query().Get("date")
	var dayStart time.Time
	if dateStr == "" {
		now := time.Now().In(loc)
		dayStart = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	} else {
		dayStart, err = time.ParseInLocation("2006-01-02", dateStr, loc)
		if err != nil {
			ErrorJSON(w, http.StatusBadRequest, "date must be YYYY-MM-DD", nil)
			return
		}
	}
	dayEnd := dayStart.Add(24 * time.Hour)

	rows, err := d.Q.ListAppointmentsByProviderOnDate(r.Context(), gen.ListAppointmentsByProviderOnDateParams{
		ProviderID:  providerID,
		StartTime:   dayStart,
		StartTime_2: dayEnd,
	})
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "failed to load provider schedule", nil)
		return
	}

	JSON(w, http.StatusOK, struct {
		ProviderID   int64       `json:"provider_id"`
		Date         string      `json:"date"`
		Appointments interface{} `json:"appointments"`
	}{
		ProviderID:   providerID,
		Date:         dayStart.Format("2006-01-02"),
		Appointments: rows,
	})
}

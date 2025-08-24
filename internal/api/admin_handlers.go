package api

import (
	"net/http"
	"time"

	"github.com/justanamir/medappoint/internal/config"
	"github.com/justanamir/medappoint/internal/db/gen"
)

type AdminDeps struct {
	Cfg config.Config
	Q   *gen.Queries
}

// GET /v1/admin/appointments?date=YYYY-MM-DD
func (d AdminDeps) ListDayAppointments(w http.ResponseWriter, r *http.Request) {
	_, ok := UserIDFromCtx(r)
	if !ok {
		ErrorJSON(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	role, _ := RoleFromCtx(r)
	if role != "admin" {
		ErrorJSON(w, http.StatusForbidden, "forbidden", nil)
		return
	}

	loc, _ := time.LoadLocation("Asia/Kuala_Lumpur")
	dateStr := r.URL.Query().Get("date")

	var dayStart time.Time
	var err error
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

	rows, err := d.Q.ListAllAppointmentsOnDate(r.Context(), gen.ListAllAppointmentsOnDateParams{
		StartTime:   dayStart,
		StartTime_2: dayEnd,
	})
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "failed to load appointments", nil)
		return
	}

	JSON(w, http.StatusOK, struct {
		Date         string      `json:"date"`
		Appointments interface{} `json:"appointments"`
	}{
		Date:         dayStart.Format("2006-01-02"),
		Appointments: rows,
	})
}

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/justanamir/medappoint/internal/db/gen"
	"github.com/justanamir/medappoint/internal/slots"
)

type AppointmentDeps struct {
	Q *gen.Queries
}

type createApptReq struct {
	ProviderID int64  `json:"provider_id"`
	PatientID  int64  `json:"patient_id"`
	ServiceID  int64  `json:"service_id"`
	StartTime  string `json:"start_time"` // RFC3339, e.g. "2025-08-25T09:00:00+08:00"
	Notes      string `json:"notes"`
}

func (d AppointmentDeps) CreateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req createApptReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid json", nil)
		return
	}
	if req.ProviderID <= 0 || req.PatientID <= 0 || req.ServiceID <= 0 || req.StartTime == "" {
		ErrorJSON(w, http.StatusBadRequest, "missing required fields", "provider_id, patient_id, service_id, start_time")
		return
	}

	// Parse start time (must include timezone offset, e.g. +08:00)
	start, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "start_time must be RFC3339 (with timezone)", "e.g. 2025-08-25T09:00:00+08:00")
		return
	}

	// Load service to get duration
	svc, err := d.Q.GetService(ctx, req.ServiceID)
	if err != nil {
		ErrorJSON(w, http.StatusNotFound, "service not found", nil)
		return
	}
	if svc.DurationMin <= 0 {
		ErrorJSON(w, http.StatusBadRequest, "invalid service duration", nil)
		return
	}
	end := start.Add(time.Duration(svc.DurationMin) * time.Minute)

	// Confirm provider exists (also gives us clinic_id)
	prov, err := d.Q.GetProvider(ctx, req.ProviderID)
	if err != nil {
		ErrorJSON(w, http.StatusNotFound, "provider not found", nil)
		return
	}

	// Basic “not in the past” check
	now := time.Now().In(start.Location())
	if !start.After(now) {
		ErrorJSON(w, http.StatusBadRequest, "cannot book a past time", nil)
		return
	}

	// Ensure request fits inside an availability window for that weekday
	// Convert Go weekday (0=Sun..6=Sat) => DB weekday (1=Mon..7=Sun)
	goWD := int(start.Weekday()) // 0..6
	dbWD := ((goWD + 6) % 7) + 1 // 1..7
	avRows, err := d.Q.GetProviderWeekdayAvailability(ctx, gen.GetProviderWeekdayAvailabilityParams{
		ProviderID: req.ProviderID,
		Weekday:    int32(dbWD),
	})
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "failed to load availability", nil)
		return
	}
	loc := start.Location()
	dayStart := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, loc)
	withinAvail := false
	for _, a := range avRows {
		ws, we, aerr := windowTimes(dayStart, a.StartHhmm, a.EndHhmm)
		if aerr != nil {
			continue
		}
		// slot must be fully contained in a window: [start,end) ⊆ [ws,we]
		if (start.Equal(ws) || start.After(ws)) && !end.After(we) {
			withinAvail = true
			break
		}
	}
	if !withinAvail {
		ErrorJSON(w, http.StatusBadRequest, "requested time is outside provider availability", nil)
		return
	}

	// Check existing appointments on that date for overlaps
	dayEnd := dayStart.Add(24 * time.Hour)
	appts, err := d.Q.ListProviderAppointmentsOnDate(ctx, gen.ListProviderAppointmentsOnDateParams{
		ProviderID:  req.ProviderID,
		StartTime:   dayStart,
		StartTime_2: dayEnd,
	})
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "failed to check overlaps", nil)
		return
	}
	booked := make([]slots.BookedRange, 0, len(appts))
	for _, ap := range appts {
		booked = append(booked, slots.BookedRange{
			Start: ap.StartTime.In(loc),
			End:   ap.EndTime.In(loc),
		})
	}
	if overlapsAny(start, end, booked) {
		ErrorJSON(w, http.StatusConflict, "time overlaps an existing appointment", nil)
		return
	}

	// All good — create appointment
	var notesPtr *string
	if req.Notes != "" {
		notesPtr = &req.Notes
	}
	row, err := d.Q.CreateAppointment(ctx, gen.CreateAppointmentParams{
		ClinicID:   prov.ClinicID,
		ProviderID: req.ProviderID,
		PatientID:  req.PatientID,
		ServiceID:  req.ServiceID,
		StartTime:  start,
		EndTime:    end,
		Notes:      notesPtr,
	})
	if err != nil {
		// could add pg error code handling here for uniqueness/exclusion constraint
		ErrorJSON(w, http.StatusBadRequest, "failed to create appointment", nil)
		return
	}

	JSON(w, http.StatusCreated, row)
}

// CancelHandler: DELETE /v1/appointments/{id}
// Rules:
// - Patient can cancel their own appointment
// - Provider/Admin can cancel any (simple policy for now)
// - Only 'scheduled' can be cancelled; returns 409 if already not-cancellable
func (d AppointmentDeps) CancelHandler(w http.ResponseWriter, r *http.Request) {
	// must be authenticated
	uid, ok := UserIDFromCtx(r)
	if !ok || uid <= 0 {
		ErrorJSON(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	role, _ := RoleFromCtx(r)

	// parse path param (robust: try chi param, then fallback to last path segment)
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) > 0 {
			idStr = parts[len(parts)-1]
		}
	}
	apptID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || apptID <= 0 {
		ErrorJSON(w, http.StatusBadRequest, "invalid appointment id", nil)
		return
	}

	ctx := r.Context()

	// load appointment
	appt, err := d.Q.GetAppointment(ctx, apptID)
	if err != nil {
		ErrorJSON(w, http.StatusNotFound, "appointment not found", nil)
		return
	}

	// if patient, ensure it's theirs
	if role == "patient" {
		// map user -> patient_id
		p, err := d.Q.GetPatientByUserID(ctx, uid)
		if err != nil {
			ErrorJSON(w, http.StatusForbidden, "patient profile not found", nil)
			return
		}
		if p.ID != appt.PatientID {
			ErrorJSON(w, http.StatusForbidden, "forbidden", nil)
			return
		}
	}

	// don't allow cancelling past appointments
	now := time.Now().In(appt.StartTime.Location())
	if !appt.StartTime.After(now) {
		ErrorJSON(w, http.StatusBadRequest, "cannot cancel past/ongoing appointment", nil)
		return
	}

	// perform cancellation
	row, err := d.Q.CancelAppointment(ctx, apptID)
	if err != nil {
		// If status wasn't 'scheduled', our WHERE matched 0 rows and sqlc will surface an error.
		ErrorJSON(w, http.StatusConflict, "cannot cancel appointment (maybe already cancelled?)", nil)
		return
	}

	JSON(w, http.StatusOK, row)
}

// ---- local helpers (mirror those in slots but local to this file) ----

func windowTimes(dayStart time.Time, startHHMM, endHHMM string) (time.Time, time.Time, error) {
	sm, err := parseHHMM(startHHMM)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	em, err := parseHHMM(endHHMM)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	ws := dayStart.Add(time.Duration(sm) * time.Minute)
	we := dayStart.Add(time.Duration(em) * time.Minute)
	if !we.After(ws) {
		return time.Time{}, time.Time{}, fmt.Errorf("end must be after start")
	}
	return ws, we, nil
}

func parseHHMM(hhmm string) (int, error) {
	if len(hhmm) != 5 || hhmm[2] != ':' {
		return 0, fmt.Errorf("invalid HH:MM: %q", hhmm)
	}
	h := int(hhmm[0]-'0')*10 + int(hhmm[1]-'0')
	m := int(hhmm[3]-'0')*10 + int(hhmm[4]-'0')
	if h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, fmt.Errorf("invalid HH:MM range: %q", hhmm)
	}
	return h*60 + m, nil
}

func overlapsAny(s, e time.Time, booked []slots.BookedRange) bool {
	for _, b := range booked {
		if s.Before(b.End) && e.After(b.Start) {
			return true
		}
	}
	return false
}

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.ProviderID <= 0 || req.PatientID <= 0 || req.ServiceID <= 0 || req.StartTime == "" {
		http.Error(w, "provider_id, patient_id, service_id, start_time are required", http.StatusBadRequest)
		return
	}

	// Parse start time (must include timezone offset, e.g. +08:00)
	start, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		http.Error(w, "start_time must be RFC3339, e.g. 2025-08-25T09:00:00+08:00", http.StatusBadRequest)
		return
	}

	// Load service to get duration
	svc, err := d.Q.GetService(ctx, req.ServiceID)
	if err != nil {
		http.Error(w, "service not found", http.StatusNotFound)
		return
	}
	if svc.DurationMin <= 0 {
		http.Error(w, "invalid service duration", http.StatusBadRequest)
		return
	}
	end := start.Add(time.Duration(svc.DurationMin) * time.Minute)

	// Confirm provider exists (also gives us clinic_id)
	prov, err := d.Q.GetProvider(ctx, req.ProviderID)
	if err != nil {
		http.Error(w, "provider not found", http.StatusNotFound)
		return
	}

	// Basic “not in the past” check
	now := time.Now().In(start.Location())
	if !start.After(now) {
		http.Error(w, "cannot book a past time", http.StatusBadRequest)
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
		http.Error(w, "failed to load availability", http.StatusInternalServerError)
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
		http.Error(w, "requested time is outside provider availability", http.StatusBadRequest)
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
		http.Error(w, "failed to check overlaps", http.StatusInternalServerError)
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
		http.Error(w, "time overlaps an existing appointment", http.StatusConflict)
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
		http.Error(w, "failed to create appointment", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(row)
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

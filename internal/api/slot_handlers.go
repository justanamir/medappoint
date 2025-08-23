package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/justanamir/medappoint/internal/db/gen"
	"github.com/justanamir/medappoint/internal/slots"
)

type SlotDeps struct {
	Q *gen.Queries
}

// GET /v1/slots?provider_id=1&service_id=1&date=2025-08-25
func (d SlotDeps) ListSlotsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// --- parse & validate query params ---
	qp := r.URL.Query()

	pidStr := qp.Get("provider_id")
	sidStr := qp.Get("service_id")
	dateStr := qp.Get("date") // YYYY-MM-DD

	if pidStr == "" || sidStr == "" || dateStr == "" {
		http.Error(w, "provider_id, service_id, and date are required", http.StatusBadRequest)
		return
	}
	providerID, err := strconv.ParseInt(pidStr, 10, 64)
	if err != nil || providerID <= 0 {
		http.Error(w, "provider_id must be a positive integer", http.StatusBadRequest)
		return
	}
	serviceID, err := strconv.ParseInt(sidStr, 10, 64)
	if err != nil || serviceID <= 0 {
		http.Error(w, "service_id must be a positive integer", http.StatusBadRequest)
		return
	}

	// Load provider/service + inputs we need
	svc, err := d.Q.GetService(ctx, serviceID)
	if err != nil {
		http.Error(w, "service not found", http.StatusNotFound)
		return
	}
	// ensure provider exists (and to keep the door open for clinic checks later)
	if _, err := d.Q.GetProvider(ctx, providerID); err != nil {
		http.Error(w, "provider not found", http.StatusNotFound)
		return
	}

	// Use clinic/provider timezone. For now we assume Asia/Kuala_Lumpur.
	loc, _ := time.LoadLocation("Asia/Kuala_Lumpur")

	// Parse date in that location
	date, err := time.ParseInLocation("2006-01-02", dateStr, loc)
	if err != nil {
		http.Error(w, "date must be YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	// weekday mapping: our DB uses 1=Mon ... 7=Sun; Go uses 0=Sun ... 6=Sat
	goWD := int(date.Weekday())  // 0..6 (Sun..Sat)
	dbWD := ((goWD + 6) % 7) + 1 // 1..7 (Mon..Sun)

	// Fetch availability rows for that weekday
	avParams := gen.GetProviderWeekdayAvailabilityParams{
		ProviderID: providerID,
		Weekday:    int32(dbWD),
	}
	avRows, err := d.Q.GetProviderWeekdayAvailability(ctx, avParams)
	if err != nil {
		http.Error(w, "failed to load availability", http.StatusInternalServerError)
		return
	}

	av := make([]slots.AvailWindow, 0, len(avRows))
	for _, a := range avRows {
		av = append(av, slots.AvailWindow{
			StartHHMM: a.StartHhmm,
			EndHHMM:   a.EndHhmm,
		})
	}

	// Fetch existing appointments for that date (to exclude overlaps)
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, loc)
	dayEnd := dayStart.Add(24 * time.Hour)
	apParams := gen.ListProviderAppointmentsOnDateParams{
		ProviderID:  providerID,
		StartTime:   dayStart,
		StartTime_2: dayEnd, // sqlc uses _2 when the same name appears twice in the SQL
	}
	appts, err := d.Q.ListProviderAppointmentsOnDate(ctx, apParams)
	if err != nil {
		http.Error(w, "failed to load appointments", http.StatusInternalServerError)
		return
	}

	booked := make([]slots.BookedRange, 0, len(appts))
	for _, ap := range appts {
		booked = append(booked, slots.BookedRange{
			Start: ap.StartTime.In(loc),
			End:   ap.EndTime.In(loc),
		})
	}

	// Generate candidate start times
	now := time.Now().In(loc)
	slotTimes, err := slots.Generate(date, loc, int(svc.DurationMin), av, booked, now)
	if err != nil {
		http.Error(w, "failed to generate slots: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Return ISO8601 timestamps to be unambiguous
	resp := struct {
		Date       string   `json:"date"`
		ProviderID int64    `json:"provider_id"`
		ServiceID  int64    `json:"service_id"`
		Slots      []string `json:"slots"`
		Count      int      `json:"count"`
	}{
		Date:       dateStr,
		ProviderID: providerID,
		ServiceID:  serviceID,
		Slots:      make([]string, 0, len(slotTimes)),
		Count:      len(slotTimes),
	}
	for _, t := range slotTimes {
		resp.Slots = append(resp.Slots, t.Format(time.RFC3339))
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

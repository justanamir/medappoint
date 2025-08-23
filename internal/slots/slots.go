package slots

import (
	"errors"
	"sort"
	"time"
)

// AvailWindow represents a provider's daily opening window on a specific date.
// We pass HH:MM strings (24h) as stored in DB (e.g. "09:00", "17:00").
type AvailWindow struct {
	StartHHMM string
	EndHHMM   string
}

// BookedRange represents an already-booked appointment interval on that date.
type BookedRange struct {
	Start time.Time
	End   time.Time
}

// Generate returns the list of bookable START times on `date` (local to `loc`),
// using the provider's availability windows, the service duration, and
// same-day "now" cutoff (i.e., no slots in the past).
//
// Inputs:
//   - date: the calendar date you want slots for (local date in `loc`, midnight-based)
//   - loc:  time.Location for the provider/clinic timezone (e.g., Asia/Kuala_Lumpur)
//   - durationMin: service duration (minutes), > 0
//   - avails: availability windows for that weekday (e.g., 09:00–17:00)
//   - booked: existing booked ranges for that date (in `loc`)
//   - now: "current time" in `loc` (pass time.Now().In(loc)); used to hide past slots on same day
//
// Output: slice of candidate start times in ascending order.
func Generate(date time.Time, loc *time.Location, durationMin int, avails []AvailWindow, booked []BookedRange, now time.Time) ([]time.Time, error) {
	if loc == nil {
		return nil, errors.New("loc is required")
	}
	if durationMin <= 0 {
		return nil, errors.New("durationMin must be > 0")
	}

	// Normalize inputs to the target location (defensive).
	date = date.In(loc)
	now = now.In(loc)

	// Start-of-day for that date in loc.
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, loc)

	step := time.Duration(durationMin) * time.Minute
	var out []time.Time

	for _, w := range avails {
		ws, we, err := windowTimes(dayStart, loc, w.StartHHMM, w.EndHHMM)
		if err != nil {
			return nil, err
		}
		// Walk the window in increments of duration.
		for t := ws; !t.Add(step).After(we); t = t.Add(step) {
			// Hide past slots if the date is today.
			if sameYMD(t, now) && !t.After(now) {
				continue
			}
			// Candidate interval [t, t+step)
			tEnd := t.Add(step)
			if overlapsAny(t, tEnd, booked) {
				continue
			}
			out = append(out, t)
		}
	}

	// Ensure ascending order and dedupe (in case windows touch).
	sort.Slice(out, func(i, j int) bool { return out[i].Before(out[j]) })
	out = uniqTimes(out)

	return out, nil
}

// windowTimes parses HH:MM strings and returns absolute times on the given day.
func windowTimes(dayStart time.Time, loc *time.Location, startHHMM, endHHMM string) (time.Time, time.Time, error) {
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
		return time.Time{}, time.Time{}, errors.New("end must be after start")
	}
	return ws, we, nil
}

// parseHHMM converts "HH:MM" (24h) to minutes since midnight.
func parseHHMM(hhmm string) (int, error) {
	if len(hhmm) != 5 || hhmm[2] != ':' {
		return 0, errors.New("invalid HH:MM")
	}
	h := (int(hhmm[0]-'0')*10 + int(hhmm[1]-'0'))
	m := (int(hhmm[3]-'0')*10 + int(hhmm[4]-'0'))
	if h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, errors.New("invalid HH:MM range")
	}
	return h*60 + m, nil
}

func overlapsAny(s, e time.Time, booked []BookedRange) bool {
	for _, b := range booked {
		// Intervals [s,e) and [b.Start,b.End) overlap if s < b.End && e > b.Start
		if s.Before(b.End) && e.After(b.Start) {
			return true
		}
	}
	return false
}

func sameYMD(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

func uniqTimes(in []time.Time) []time.Time {
	if len(in) == 0 {
		return in
	}
	out := make([]time.Time, 0, len(in))
	out = append(out, in[0])
	for i := 1; i < len(in); i++ {
		if !in[i].Equal(in[i-1]) {
			out = append(out, in[i])
		}
	}
	return out
}

package timeline

import "time"

// tickUnit is the calendar unit a tick step is expressed in.
type tickUnit int

const (
	unitHour tickUnit = iota
	unitDay
	unitMonth
	unitYear
)

// tickStep is a candidate spacing between axis ticks.
type tickStep struct {
	unit   tickUnit
	mult   int
	approx time.Duration // rough length, used only to rank candidates
}

const day = 24 * time.Hour

// tickSteps lists candidate spacings from finest to coarsest. niceTicks picks
// the finest a single that yields at most maxTicks ticks across the span.
var tickSteps = []tickStep{
	{unitHour, 1, time.Hour},
	{unitHour, 2, 2 * time.Hour},
	{unitHour, 3, 3 * time.Hour},
	{unitHour, 6, 6 * time.Hour},
	{unitHour, 12, 12 * time.Hour},
	{unitDay, 1, day},
	{unitDay, 2, 2 * day},
	{unitDay, 7, 7 * day},
	{unitDay, 14, 14 * day},
	{unitMonth, 1, 30 * day},
	{unitMonth, 2, 61 * day},
	{unitMonth, 3, 91 * day},
	{unitMonth, 6, 182 * day},
	{unitYear, 1, 365 * day},
	{unitYear, 2, 2 * 365 * day},
	{unitYear, 5, 5 * 365 * day},
	{unitYear, 10, 10 * 365 * day},
	{unitYear, 20, 20 * 365 * day},
	{unitYear, 50, 50 * 365 * day},
	{unitYear, 100, 100 * 365 * day},
}

// niceTicks returns evenly spaced, calendar-aligned tick times between start
// and end (inclusive of any that fall in range).
func niceTicks(start, end time.Time, maxTicks int) []time.Time {
	if maxTicks < 1 || !end.After(start) {
		return nil
	}
	span := end.Sub(start)
	chosen := tickSteps[len(tickSteps)-1]
	for _, s := range tickSteps {
		if int(span/s.approx) <= maxTicks {
			chosen = s
			break
		}
	}
	var ticks []time.Time
	for t := alignUp(start, chosen); !t.After(end); t = advance(t, chosen) {
		if !t.Before(start) {
			ticks = append(ticks, t)
		}
		if len(ticks) > maxTicks+2 { // safety valve
			break
		}
	}
	return ticks
}

// alignUp rounds t up to the next boundary for the step's unit. Months align
// to a multiple of mult (so quarters land on Jan/Apr/Jul/Oct) and years align
// to a multiple of mult (so decades land on round years).
func alignUp(t time.Time, s tickStep) time.Time {
	loc := t.Location()
	switch s.unit {
	case unitHour:
		a := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, loc)
		if a.Before(t) {
			a = a.Add(time.Hour)
		}
		return a
	case unitDay:
		a := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
		if a.Before(t) {
			a = a.AddDate(0, 0, 1)
		}
		return a
	case unitMonth:
		a := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, loc)
		if a.Before(t) {
			a = a.AddDate(0, 1, 0)
		}
		for s.mult > 1 && (int(a.Month())-1)%s.mult != 0 {
			a = a.AddDate(0, 1, 0)
		}
		return a
	case unitYear:
		y := (t.Year() / s.mult) * s.mult
		a := time.Date(y, 1, 1, 0, 0, 0, 0, loc)
		if a.Before(t) {
			a = time.Date(y+s.mult, 1, 1, 0, 0, 0, 0, loc)
		}
		return a
	}
	return t
}

// advance moves t forward by a single step.
func advance(t time.Time, s tickStep) time.Time {
	switch s.unit {
	case unitHour:
		return t.Add(time.Duration(s.mult) * time.Hour)
	case unitDay:
		return t.AddDate(0, 0, s.mult)
	case unitMonth:
		return t.AddDate(0, s.mult, 0)
	case unitYear:
		return t.AddDate(s.mult, 0, 0)
	}
	return t
}

// tickFormat picks a compact label layout for the axis ticks, appropriate for
// the span. Ticks act as gridlines, so the layout stays coarse.
func tickFormat(start, end time.Time) string {
	days := end.Sub(start).Hours() / 24
	switch {
	case days <= 2:
		return "Jan 2 15:04"
	case days <= 120:
		return "Jan 2"
	case days <= 1095:
		return "Jan 2006"
	default:
		return "2006"
	}
}

// eventDateFormat picks the per-event date layout. It is generally a single
// notch more precise than the axis ticks so each event keeps its day (and
// time, for short spans) while still reading cleanly.
func eventDateFormat(start, end time.Time) string {
	days := end.Sub(start).Hours() / 24
	switch {
	case days <= 2:
		return "Jan 2 15:04"
	case days > 1095:
		return "Jan 2006"
	case start.Year() != end.Year() || days > 300:
		return "Jan 2, 2006"
	default:
		return "Jan 2"
	}
}

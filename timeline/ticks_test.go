package timeline

import (
	"testing"
	"time"
)

func mustDate(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := parseDate(s)
	if err != nil {
		t.Fatalf("parseDate(%q): %v", s, err)
	}
	return d
}

func TestNiceTicksWithinRangeAndSorted(t *testing.T) {
	spans := [][2]string{
		{"2024-01-01", "2024-12-31"}, // ~1 year
		{"2024-01-01", "2024-01-04"}, // a few days
		{"2020-01-01", "2030-01-01"}, // a decade
		{"2024-06-01", "2024-06-02"}, // one day
	}
	for _, sp := range spans {
		s, e := mustDate(t, sp[0]), mustDate(t, sp[1])
		ticks := niceTicks(s, e, 6)
		if len(ticks) == 0 {
			t.Errorf("%v..%v: got no ticks", sp[0], sp[1])
			continue
		}
		if len(ticks) > 8 {
			t.Errorf("%v..%v: got %d ticks, expected a handful", sp[0], sp[1], len(ticks))
		}
		prev := time.Time{}
		for _, tk := range ticks {
			if tk.Before(s) || tk.After(e) {
				t.Errorf("%v..%v: tick %v out of range", sp[0], sp[1], tk)
			}
			if !prev.IsZero() && !tk.After(prev) {
				t.Errorf("%v..%v: ticks not strictly increasing at %v", sp[0], sp[1], tk)
			}
			prev = tk
		}
	}
}

func TestNiceTicksEmptyWhenNoSpan(t *testing.T) {
	d := mustDate(t, "2024-01-01")
	if ticks := niceTicks(d, d, 6); ticks != nil {
		t.Errorf("expected nil ticks for zero span, got %v", ticks)
	}
}

func TestTickFormat(t *testing.T) {
	cases := []struct {
		start, end, want string
	}{
		{"2024-01-01", "2024-01-02", "Jan 2 15:04"},
		{"2024-01-01", "2024-02-15", "Jan 2"},
		{"2024-01-01", "2024-12-31", "Jan 2006"},
		{"2000-01-01", "2030-01-01", "2006"},
	}
	for _, c := range cases {
		got := tickFormat(mustDate(t, c.start), mustDate(t, c.end))
		if got != c.want {
			t.Errorf("tickFormat(%s,%s) = %q, want %q", c.start, c.end, got, c.want)
		}
	}
}

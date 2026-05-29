package timeline

import (
	"strings"
	"testing"
	"time"
)

func sampleTimeline() *Timeline {
	tl := New()
	tl.NoColor = true
	tl.Header = "# Demo"
	tl.Description = "A short description."
	tl.Start = mustDateT("2024-01-01")
	tl.End = mustDateT("2024-12-31")
	tl.Events = []Event{
		{Label: "GA launch", Date: mustDateT("2024-12-02")},
		{Label: "Kickoff", Date: mustDateT("2024-01-08")},
		{Label: "Build", Start: mustDateT("2024-03-01"), End: mustDateT("2024-08-30")},
	}
	return tl
}

// mustDateT is a non-*testing.T helper for table data above.
func mustDateT(s string) time.Time {
	d, err := parseDate(s)
	if err != nil {
		panic(err)
	}
	return d
}

func TestColFor(t *testing.T) {
	rc := renderCtx{
		s:     mustDateT("2024-01-01"),
		e:     mustDateT("2024-01-11"),
		width: 11,
	}
	if c := colFor(rc.s, rc); c != 0 {
		t.Errorf("start column = %d, want 0", c)
	}
	if c := colFor(rc.e, rc); c != rc.width-1 {
		t.Errorf("end column = %d, want %d", c, rc.width-1)
	}
	if c := colFor(mustDateT("2024-01-06"), rc); c != 5 {
		t.Errorf("midpoint column = %d, want 5", c)
	}
	// Out-of-range times clamp into the axis.
	if c := colFor(mustDateT("2023-01-01"), rc); c != 0 {
		t.Errorf("before-start column = %d, want 0", c)
	}
	if c := colFor(mustDateT("2025-01-01"), rc); c != rc.width-1 {
		t.Errorf("after-end column = %d, want %d", c, rc.width-1)
	}
}

func TestRenderStructure(t *testing.T) {
	out := sampleTimeline().String()

	for _, want := range []string{
		"# Demo",      // header
		"A short",     // description
		"→",           // range caption arrow
		"Jan 2024",    // start tick label (Jan 2006 format)
		"Dec 2024",    // end tick label
		"├", "┤", "─", // axis
		"●", // point marker
		"█", // duration bar
		"Kickoff", "Build", "GA launch",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("render output missing %q\n---\n%s", want, out)
		}
	}
}

func TestRenderChronologicalOrder(t *testing.T) {
	out := sampleTimeline().String()
	ki := strings.Index(out, "Kickoff")
	bi := strings.Index(out, "Build")
	gi := strings.Index(out, "GA launch")
	if !(ki < bi && bi < gi) {
		t.Errorf("events not in chronological order: Kickoff@%d Build@%d GA@%d", ki, bi, gi)
	}
}

func TestNoColorHasNoEscapes(t *testing.T) {
	if strings.Contains(sampleTimeline().String(), "\x1b") {
		t.Error("NoColor output should contain no ANSI escapes")
	}
}

func TestColorHasEscapes(t *testing.T) {
	tl := sampleTimeline()
	tl.NoColor = false
	if !strings.Contains(tl.String(), "\x1b[") {
		t.Error("colored output should contain ANSI escapes")
	}
}

func TestEmptyTimelineIsGraceful(t *testing.T) {
	tl := New()
	tl.NoColor = true
	out := tl.String()
	if !strings.Contains(out, DefaultHeader) {
		t.Errorf("empty timeline should still show default header, got:\n%s", out)
	}
	if !strings.Contains(out, "(no events)") {
		t.Errorf("empty timeline should note it is empty, got:\n%s", out)
	}
}

func TestBoundsDerivedFromEvents(t *testing.T) {
	tl := New()
	tl.NoColor = true
	tl.Events = []Event{
		{Label: "a", Date: mustDateT("2024-02-01")},
		{Label: "b", Date: mustDateT("2024-08-01")},
	}
	s, e, ok := tl.bounds()
	if !ok {
		t.Fatal("bounds not ok")
	}
	// Derived bounds are padded outward, so they straddle the events.
	if !s.Before(mustDateT("2024-02-01")) {
		t.Errorf("derived start %v should be padded before first event", s)
	}
	if !e.After(mustDateT("2024-08-01")) {
		t.Errorf("derived end %v should be padded after last event", e)
	}
}

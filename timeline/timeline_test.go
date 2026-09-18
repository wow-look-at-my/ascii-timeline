package timeline

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.Equal(t, 0, colFor(rc.s, rc))
	assert.Equal(t, rc.width-1, colFor(rc.e, rc))
	assert.Equal(t, 5, colFor(mustDateT("2024-01-06"), rc))

	// Out-of-range times clamp into the axis.
	assert.Equal(t, 0, colFor(mustDateT("2023-01-01"), rc))
	assert.Equal(t, rc.width-1, colFor(mustDateT("2025-01-01"), rc))
}

func TestRenderStructure(t *testing.T) {
	out := sampleTimeline().String()

	for _, want := range []string{
		"# Demo",       // header
		"A short",      // description
		"→",            // range caption arrow
		"Jan 2024",     // a left-aligned axis tick label
		"Nov 2024",     // a later axis tick label
		"Dec 31, 2024", // exact end date shown in the caption
		"Jan 8, 2024",  // event date keeps the day (finer than ticks)
		"├", "┤", "─",  // axis
		"●", // point marker
		"█", // duration bar
		"Kickoff", "Build", "GA launch",
	} {
		assert.Contains(t, out, want)

	}
}

func TestRenderChronologicalOrder(t *testing.T) {
	out := sampleTimeline().String()
	ki := strings.Index(out, "Kickoff")
	bi := strings.Index(out, "Build")
	gi := strings.Index(out, "GA launch")
	assert.True(t, (ki < bi && bi < gi))

}

func TestNoColorHasNoEscapes(t *testing.T) {
	assert.NotContains(t, sampleTimeline().String(), "\x1b")

}

func TestColorHasEscapes(t *testing.T) {
	tl := sampleTimeline()
	tl.NoColor = false
	assert.Contains(t, tl.String(), "\x1b[")

}

func TestEmptyTimelineIsGraceful(t *testing.T) {
	tl := New()
	tl.NoColor = true
	out := tl.String()
	assert.Contains(t, out, DefaultHeader)

	assert.Contains(t, out, "(no events)")

}

func TestBoundsDerivedFromEvents(t *testing.T) {
	tl := New()
	tl.NoColor = true
	tl.Events = []Event{
		{Label: "a", Date: mustDateT("2024-02-01")},
		{Label: "b", Date: mustDateT("2024-08-01")},
	}
	s, e, ok := tl.bounds()
	require.True(t, ok)

	// Derived bounds are padded outward, so they straddle the events.
	assert.True(t, s.Before(mustDateT("2024-02-01")))

	assert.True(t, e.After(mustDateT("2024-08-01")))

}

// todayAt returns a midnight time at the given day offset from now.
func todayAt(offset int) time.Time {
	now := time.Now()
	d := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return d.AddDate(0, 0, offset)
}

func TestTodayMarkerAppearsWhenInRange(t *testing.T) {
	tl := New()
	tl.NoColor = true
	tl.Start = todayAt(-30)
	tl.End = todayAt(30)
	tl.Events = []Event{
		{Label: "past", Date: todayAt(-20)},
		{Label: "future", Date: todayAt(20)},
	}
	out := tl.String()
	assert.Contains(t, out, "today")
	assert.Contains(t, out, "▼")
	assert.Contains(t, out, "│") // vertical line through event rows
}

func TestTodayMarkerAbsentWhenOutOfRange(t *testing.T) {
	tl := New()
	tl.NoColor = true
	tl.Start = mustDateT("2000-01-01")
	tl.End = mustDateT("2000-12-31")
	tl.Events = []Event{{Label: "e", Date: mustDateT("2000-06-01")}}
	out := tl.String()
	assert.NotContains(t, out, "today")
	assert.NotContains(t, out, "▼")
}

func TestTodayMarkerLabelNoRemnants(t *testing.T) {
	// Today is near the left edge so "today" can overlap a tick label.
	// Verify no digit from the overwritten tick bleeds through.
	tl := New()
	tl.NoColor = true
	tl.Events = []Event{{Label: "e", Date: todayAt(7)}}
	out := tl.String()
	assert.Contains(t, out, "today")
	for _, d := range "0123456789" {
		assert.NotContains(t, out, "today"+string(d))
	}
}

func TestBoundsExtendedForTodayWithinTwoWeeks(t *testing.T) {
	tl := New()
	tl.Events = []Event{{Label: "e", Date: todayAt(-7)}}
	_, e, ok := tl.bounds()
	require.True(t, ok)
	assert.False(t, e.Before(todayAt(0)), "end should reach today when event is within 2 weeks")
	assert.Contains(t, tl.String(), "today")

	tl2 := New()
	tl2.Events = []Event{{Label: "e", Date: todayAt(7)}}
	s2, _, ok2 := tl2.bounds()
	require.True(t, ok2)
	assert.False(t, todayAt(0).Before(s2), "start should reach today when event is within 2 weeks")
	assert.Contains(t, tl2.String(), "today")
}

func TestBoundsNotExtendedBeyondTwoWeeks(t *testing.T) {
	tl := New()
	tl.NoColor = true
	tl.Events = []Event{{Label: "e", Date: todayAt(-21)}}
	assert.NotContains(t, tl.String(), "today")

	tl2 := New()
	tl2.NoColor = true
	tl2.Events = []Event{{Label: "e", Date: todayAt(21)}}
	assert.NotContains(t, tl2.String(), "today")
}

func TestBoundsNotExtendedWhenExplicit(t *testing.T) {
	// Explicit start/end that exclude today should never be stretched.
	tl := New()
	tl.NoColor = true
	tl.Start = todayAt(5)
	tl.End = todayAt(30)
	tl.Events = []Event{{Label: "e", Date: todayAt(10)}}
	assert.NotContains(t, tl.String(), "today")
}

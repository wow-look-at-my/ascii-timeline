// Package timeline renders a horizontal, annotated timeline as styled text.
//
// A Timeline has an optional header and description followed by a date axis
// running left (start) to right (end). Events are drawn beneath the axis,
// sorted chronologically: point events as a single marker, duration events as
// a bar spanning their range. Output uses Unicode box-drawing characters and
// ANSI color by default; set NoColor to emit plain text.
//
// Timelines are usually built from JSON (see Parse) so an LLM or a human can
// describe one declaratively, but the struct can also be assembled directly.
package timeline

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
)

// Defaults applied by New and by Parse when a field is omitted.
const (
	DefaultHeader = "# Timeline"
	DefaultWidth  = 72
	minWidth      = 24
)

// Event is a single annotation on the timeline. It is a point event when only
// Date is set, or a duration event when both Start and End are set.
type Event struct {
	Label       string
	Description string

	Date  time.Time // point event
	Start time.Time // duration event start
	End   time.Time // duration event end

	// Color is an optional style spec (e.g. "red", "bold cyan"). When empty a
	// color is assigned automatically from a palette.
	Color string
}

func (e Event) isDuration() bool { return !e.Start.IsZero() && !e.End.IsZero() }

// at returns the event's leftmost time (its position on the axis).
func (e Event) at() time.Time {
	if e.isDuration() {
		return e.Start
	}
	return e.Date
}

// endAt returns the event's rightmost time.
func (e Event) endAt() time.Time {
	if e.isDuration() {
		return e.End
	}
	return e.Date
}

// Timeline is a renderable timeline. Every field is optional; New supplies the
// documented defaults.
type Timeline struct {
	Header      string
	Description string
	Start       time.Time // left edge; derived from events when zero
	End         time.Time // right edge; derived from events when zero
	Width       int       // axis width in columns
	Events      []Event
	NoColor     bool
}

// New returns a Timeline preloaded with the default header and width.
func New() *Timeline {
	return &Timeline{Header: DefaultHeader, Width: DefaultWidth}
}

// renderCtx carries the resolved bounds and formats through the draw helpers.
type renderCtx struct {
	s, e   time.Time
	width  int
	tickF  string // axis tick label layout
	eventF string // per-event date label layout
}

// palette supplies distinct colors to events that don't specify one.
var palette = []string{
	"cyan", "green", "yellow", "magenta", "blue", "red",
	"brightcyan", "brightgreen", "brightyellow", "brightmagenta",
}

// style wraps s in the given ANSI spec unless color is disabled.
func (t *Timeline) style(s, spec string) string {
	if t.NoColor || spec == "" {
		return s
	}
	if open := ansiOpen(spec); open != "" {
		return open + s + reset
	}
	return s
}

// Render writes the timeline to w.
func (t *Timeline) Render(w io.Writer) error {
	width := t.Width
	if width <= 0 {
		width = DefaultWidth
	}
	if width < minWidth {
		width = minWidth
	}

	var out strings.Builder

	if t.Header != "" {
		out.WriteString(t.style(t.Header, "bold"))
		out.WriteByte('\n')
	}
	if t.Description != "" {
		for _, ln := range wrapText(t.Description, width) {
			out.WriteString(ln)
			out.WriteByte('\n')
		}
	}

	s, e, ok := t.bounds()
	if !ok {
		if t.Header != "" || t.Description != "" {
			out.WriteByte('\n')
		}
		out.WriteString(t.style("(no events)", "dim"))
		out.WriteByte('\n')
		_, err := io.WriteString(w, out.String())
		return err
	}

	if t.Header != "" || t.Description != "" {
		out.WriteByte('\n')
	}

	rc := renderCtx{
		s:      s,
		e:      e,
		width:  width,
		tickF:  tickFormat(s, e),
		eventF: eventDateFormat(s, e),
	}

	out.WriteString(t.style(formatRange(s, e), "dim"))
	out.WriteByte('\n')

	cv := newCanvas()
	t.drawAxis(cv, rc)
	t.drawEvents(cv, rc)
	out.WriteString(cv.render(t.NoColor))

	_, err := io.WriteString(w, out.String())
	return err
}

// String renders the timeline to a string (convenience wrapper around Render).
func (t *Timeline) String() string {
	var b strings.Builder
	_ = t.Render(&b)
	return b.String()
}

// bounds resolves the effective [start, end] window. When an edge is not set
// explicitly it is derived from the events and padded slightly so markers
// aren't glued to the border. Returns ok=false when there is nothing to draw.
func (t *Timeline) bounds() (start, end time.Time, ok bool) {
	sProvided := !t.Start.IsZero()
	eProvided := !t.End.IsZero()
	start, end = t.Start, t.End

	var minT, maxT time.Time
	for _, ev := range t.Events {
		for _, d := range [2]time.Time{ev.at(), ev.endAt()} {
			if d.IsZero() {
				continue
			}
			if minT.IsZero() || d.Before(minT) {
				minT = d
			}
			if maxT.IsZero() || d.After(maxT) {
				maxT = d
			}
		}
	}
	if !sProvided {
		start = minT
	}
	if !eProvided {
		end = maxT
	}
	if start.IsZero() && end.IsZero() {
		return time.Time{}, time.Time{}, false
	}
	if start.IsZero() {
		start = end
	}
	if end.IsZero() {
		end = start
	}
	if !end.After(start) {
		end = start.Add(day)
	}

	pad := time.Duration(float64(end.Sub(start)) * 0.04)
	if pad <= 0 {
		pad = time.Hour
	}
	if !sProvided {
		start = start.Add(-pad)
	}
	if !eProvided {
		end = end.Add(pad)
	}
	return start, end, true
}

// colFor maps a time to a column in [0, width-1].
func colFor(d time.Time, rc renderCtx) int {
	frac := float64(d.Sub(rc.s)) / float64(rc.e.Sub(rc.s))
	c := int(frac*float64(rc.width-1) + 0.5)
	if c < 0 {
		c = 0
	}
	if c > rc.width-1 {
		c = rc.width - 1
	}
	return c
}

// drawAxis renders the tick labels (row 0) and the axis line (row 1). Labels
// are left-aligned at their tick column; the exact start/end dates live in the
// caption above, so the axis stays uncluttered.
func (t *Timeline) drawAxis(cv *canvas, rc renderCtx) {
	const labelRow, axisRow = 0, 1

	for x := 0; x < rc.width; x++ {
		cv.set(x, axisRow, '─', "dim")
	}
	cv.set(0, axisRow, '├', "dim")
	cv.set(rc.width-1, axisRow, '┤', "dim")

	occupied := make([]bool, rc.width)
	// fits reports whether a label of n runes left-aligned at col fits without
	// running off the axis or colliding with an earlier label (plus a gap).
	fits := func(col, n int) bool {
		if col < 0 || col+n+1 > rc.width {
			return false
		}
		for i := col; i < col+n+1; i++ {
			if occupied[i] {
				return false
			}
		}
		return true
	}

	for _, tk := range niceTicks(rc.s, rc.e, 6) {
		col := colFor(tk, rc)
		if col > 0 && col < rc.width-1 {
			cv.set(col, axisRow, '┴', "dim")
		}
		label := tk.Format(rc.tickF)
		n := len([]rune(label))
		if !fits(col, n) {
			continue
		}
		cv.puts(col, labelRow, label, "dim")
		for i := col; i <= col+n; i++ {
			occupied[i] = true
		}
	}
}

// drawEvents renders each event on its own row beneath the axis, in
// chronological order.
func (t *Timeline) drawEvents(cv *canvas, rc renderCtx) {
	evs := append([]Event(nil), t.Events...)
	sort.SliceStable(evs, func(i, j int) bool {
		return evs[i].at().Before(evs[j].at())
	})

	row := 2 // labels=0, axis=1
	for i, ev := range evs {
		spec := ev.Color
		if spec == "" {
			spec = palette[i%len(palette)]
		}
		startCol := colFor(ev.at(), rc)

		var labelCol int
		if ev.isDuration() {
			endCol := colFor(ev.endAt(), rc)
			if endCol < startCol {
				endCol = startCol
			}
			for x := startCol; x <= endCol; x++ {
				cv.set(x, row, '█', spec)
			}
			labelCol = endCol + 2
		} else {
			cv.set(startCol, row, '●', spec)
			labelCol = startCol + 2
		}
		t.drawEventLabel(cv, row, labelCol, ev, spec, rc)
		row++
	}
}

// drawEventLabel writes "<label> (<date>) — <description>" after the marker.
func (t *Timeline) drawEventLabel(cv *canvas, row, col int, ev Event, spec string, rc renderCtx) {
	x := col
	if ev.Label != "" {
		x = cv.puts(x, row, ev.Label, spec)
		x++
	}
	var dateStr string
	if ev.isDuration() {
		dateStr = "(" + ev.Start.Format(rc.eventF) + " – " + ev.End.Format(rc.eventF) + ")"
	} else {
		dateStr = "(" + ev.Date.Format(rc.eventF) + ")"
	}
	x = cv.puts(x, row, dateStr, "dim")
	if ev.Description != "" {
		x++
		cv.puts(x, row, "— "+ev.Description, "dim")
	}
}

// formatRange produces the dim caption above the axis, e.g.
// "Jan 1, 2024  →  Dec 31, 2024   (12 months)".
func formatRange(s, e time.Time) string {
	f := "Jan 2, 2006"
	if e.Sub(s) < 48*time.Hour {
		f = "Jan 2, 2006 15:04"
	}
	return fmt.Sprintf("%s  →  %s   (%s)", s.Format(f), e.Format(f), humanDur(s, e))
}

// humanDur describes the span between s and e in the largest sensible unit.
func humanDur(s, e time.Time) string {
	d := e.Sub(s)
	days := d.Hours() / 24
	switch {
	case d < 48*time.Hour:
		return plural(int(d.Hours()+0.5), "hour")
	case days < 60:
		return plural(int(days+0.5), "day")
	case days < 730:
		return plural(int(days/30.0+0.5), "month")
	default:
		return plural(int(days/365.0+0.5), "year")
	}
}

func plural(n int, unit string) string {
	if n == 1 {
		return "1 " + unit
	}
	return fmt.Sprintf("%d %ss", n, unit)
}

// wrapText word-wraps s to width columns.
func wrapText(s string, width int) []string {
	if width < minWidth {
		width = minWidth
	}
	words := strings.Fields(s)
	if len(words) == 0 {
		return nil
	}
	var lines []string
	cur := words[0]
	for _, w := range words[1:] {
		if len([]rune(cur))+1+len([]rune(w)) > width {
			lines = append(lines, cur)
			cur = w
			continue
		}
		cur += " " + w
	}
	return append(lines, cur)
}

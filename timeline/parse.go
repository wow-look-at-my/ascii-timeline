package timeline

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

// jsonEvent is the wire form of an Event.
type jsonEvent struct {
	Label       string `json:"label"`
	Description string `json:"description"`
	Date        string `json:"date"`
	Start       string `json:"start"`
	End         string `json:"end"`
	Color       string `json:"color"`
}

// jsonDoc is the wire form of a Timeline. Header is a pointer so an omitted
// field (default header) can be told apart from an explicit empty string
// (no header).
type jsonDoc struct {
	Header      *string     `json:"header"`
	Description string      `json:"description"`
	Start       string      `json:"start"`
	End         string      `json:"end"`
	Width       int         `json:"width"`
	Events      []jsonEvent `json:"events"`
}

// dateLayouts are tried in order; the first that parses wins. The list favors
// unambiguous ISO-style layouts before locale-specific ones.
var dateLayouts = []string{
	time.RFC3339,
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02 15:04",
	"2006-01-02T15:04",
	"2006-01-02",
	"2006/01/02",
	"01/02/2006",
	"02 Jan 2006",
	"Jan 2, 2006",
	"January 2, 2006",
	"Jan 2006",
	"January 2006",
	"2006-01",
	"2006",
}

// parseDate accepts a range of common date/datetime formats. Date-only inputs
// are interpreted at midnight UTC.
func parseDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	for _, layout := range dateLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("could not parse date %q (try e.g. 2006-01-02, \"Jan 2, 2006\", or an RFC3339 timestamp)", s)
}

// Parse reads a JSON timeline document from r.
func Parse(r io.Reader) (*Timeline, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return ParseBytes(b)
}

// ParseBytes decodes a JSON timeline document. Defaults are applied for any
// omitted field, so the smallest valid document is a single event.
func ParseBytes(b []byte) (*Timeline, error) {
	var doc jsonDoc
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	tl := New()
	if doc.Header != nil {
		tl.Header = *doc.Header
	}
	if doc.Description != "" {
		tl.Description = doc.Description
	}
	if doc.Width != 0 {
		tl.Width = doc.Width
	}

	var err error
	if doc.Start != "" {
		if tl.Start, err = parseDate(doc.Start); err != nil {
			return nil, fmt.Errorf("start: %w", err)
		}
	}
	if doc.End != "" {
		if tl.End, err = parseDate(doc.End); err != nil {
			return nil, fmt.Errorf("end: %w", err)
		}
	}

	for i, je := range doc.Events {
		ev := Event{Label: je.Label, Description: je.Description, Color: je.Color}
		who := fmt.Sprintf("event %d (%q)", i, je.Label)

		if je.Date != "" {
			if ev.Date, err = parseDate(je.Date); err != nil {
				return nil, fmt.Errorf("%s date: %w", who, err)
			}
		}
		if je.Start != "" {
			if ev.Start, err = parseDate(je.Start); err != nil {
				return nil, fmt.Errorf("%s start: %w", who, err)
			}
		}
		if je.End != "" {
			if ev.End, err = parseDate(je.End); err != nil {
				return nil, fmt.Errorf("%s end: %w", who, err)
			}
		}

		switch {
		case !ev.Date.IsZero():
			// point event, ok
		case !ev.Start.IsZero() && !ev.End.IsZero():
			if ev.End.Before(ev.Start) {
				return nil, fmt.Errorf("%s: end is before start", who)
			}
		default:
			return nil, fmt.Errorf("%s: needs either \"date\" (a point event) or both \"start\" and \"end\" (a duration event)", who)
		}

		tl.Events = append(tl.Events, ev)
	}

	return tl, nil
}

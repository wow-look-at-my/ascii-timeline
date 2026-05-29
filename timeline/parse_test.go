package timeline

import (
	"strings"
	"testing"
)

func TestParseDefaultsHeader(t *testing.T) {
	tl, err := ParseBytes([]byte(`{"events":[{"label":"x","date":"2024-01-01"}]}`))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if tl.Header != DefaultHeader {
		t.Errorf("header = %q, want default %q", tl.Header, DefaultHeader)
	}
	if len(tl.Events) != 1 || tl.Events[0].Date.IsZero() {
		t.Fatalf("expected one point event, got %+v", tl.Events)
	}
}

func TestParseExplicitEmptyHeader(t *testing.T) {
	tl, err := ParseBytes([]byte(`{"header":"","events":[{"label":"x","date":"2024"}]}`))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if tl.Header != "" {
		t.Errorf("explicit empty header should suppress default, got %q", tl.Header)
	}
}

func TestParseCustomHeader(t *testing.T) {
	tl, err := ParseBytes([]byte(`{"header":"# Hi","events":[{"label":"x","date":"2024"}]}`))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if tl.Header != "# Hi" {
		t.Errorf("header = %q, want %q", tl.Header, "# Hi")
	}
}

func TestParseDateFormats(t *testing.T) {
	cases := map[string]struct{ y, m, d int }{
		"2024-03-15":           {2024, 3, 15},
		"2024/03/15":           {2024, 3, 15},
		"Jan 2, 2006":          {2006, 1, 2},
		"03/15/2024":           {2024, 3, 15},
		"2024-03":              {2024, 3, 1},
		"2024":                 {2024, 1, 1},
		"2024-03-15T09:30:00Z": {2024, 3, 15},
		"15 Mar 2024":          {2024, 3, 15},
	}
	for in, want := range cases {
		got, err := parseDate(in)
		if err != nil {
			t.Errorf("parseDate(%q): %v", in, err)
			continue
		}
		if got.Year() != want.y || int(got.Month()) != want.m || got.Day() != want.d {
			t.Errorf("parseDate(%q) = %v, want %04d-%02d-%02d", in, got, want.y, want.m, want.d)
		}
	}
}

func TestParseBadDate(t *testing.T) {
	if _, err := parseDate("not-a-date"); err == nil {
		t.Error("expected error for unparseable date")
	}
}

func TestParseEventValidation(t *testing.T) {
	cases := []struct {
		name string
		json string
	}{
		{"no date or range", `{"events":[{"label":"x"}]}`},
		{"end before start", `{"events":[{"label":"x","start":"2024-05-01","end":"2024-01-01"}]}`},
		{"bad json", `{`},
	}
	for _, c := range cases {
		if _, err := ParseBytes([]byte(c.json)); err == nil {
			t.Errorf("%s: expected error, got nil", c.name)
		}
	}
}

func TestParseUnknownFieldsTolerated(t *testing.T) {
	// LLMs may add extra keys; we should not choke on them.
	tl, err := ParseBytes([]byte(`{"title":"ignored","events":[{"label":"x","date":"2024","note":"extra"}]}`))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(tl.Events) != 1 {
		t.Fatalf("want 1 event, got %d", len(tl.Events))
	}
}

func TestParseRejectsBadColorlessNothing(t *testing.T) {
	_, err := ParseBytes([]byte(`{"events":[{"label":"only start","start":"2024-01-01"}]}`))
	if err == nil || !strings.Contains(err.Error(), "start") {
		t.Errorf("expected a 'needs date or start+end' style error, got %v", err)
	}
}

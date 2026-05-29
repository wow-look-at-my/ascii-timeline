package timeline

import (
	"github.com/wow-look-at-my/testify/assert"
	"github.com/wow-look-at-my/testify/require"
	"strings"
	"testing"
)

func TestParseDefaultsHeader(t *testing.T) {
	tl, err := ParseBytes([]byte(`{"events":[{"label":"x","date":"2024-01-01"}]}`))
	require.Nil(t, err)

	assert.Equal(t, DefaultHeader, tl.Header)

	require.False(t, len(tl.Events) != 1 || tl.Events[0].Date.IsZero())

}

func TestParseExplicitEmptyHeader(t *testing.T) {
	tl, err := ParseBytes([]byte(`{"header":"","events":[{"label":"x","date":"2024"}]}`))
	require.Nil(t, err)

	assert.Equal(t, "", tl.Header)

}

func TestParseCustomHeader(t *testing.T) {
	tl, err := ParseBytes([]byte(`{"header":"# Hi","events":[{"label":"x","date":"2024"}]}`))
	require.Nil(t, err)

	assert.Equal(t, "# Hi", tl.Header)

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
		assert.Nil(t, err)

		assert.False(t, got.Year() != want.y || int(got.Month()) != want.m || got.Day() != want.d)

	}
}

func TestParseBadDate(t *testing.T) {
	_, err := parseDate("not-a-date")
	assert.NotNil(t, err)

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
		_, err := ParseBytes([]byte(c.json))
		assert.NotNil(t, err)

	}
}

func TestParseUnknownFieldsTolerated(t *testing.T) {
	// LLMs may add extra keys; we should not choke on them.
	tl, err := ParseBytes([]byte(`{"title":"ignored","events":[{"label":"x","date":"2024","note":"extra"}]}`))
	require.Nil(t, err)

	require.Equal(t, 1, len(tl.Events))

}

func TestParseRejectsBadColorlessNothing(t *testing.T) {
	_, err := ParseBytes([]byte(`{"events":[{"label":"only start","start":"2024-01-01"}]}`))
	assert.False(t, err == nil || !strings.Contains(err.Error(), "start"))

}

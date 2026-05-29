package timeline

import (
	"github.com/wow-look-at-my/testify/assert"
	"github.com/wow-look-at-my/testify/require"
	"testing"
	"time"
)

func mustDate(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := parseDate(s)
	require.Nil(t, err)

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
		assert.NotEqual(t, 0, len(ticks))

		assert.LessOrEqual(t, len(ticks), 8)

		prev := time.Time{}
		for _, tk := range ticks {
			assert.False(t, tk.Before(s) || tk.After(e))

			assert.False(t, !prev.IsZero() && !tk.After(prev))

			prev = tk
		}
	}
}

func TestNiceTicksEmptyWhenNoSpan(t *testing.T) {
	d := mustDate(t, "2024-01-01")
	assert.Empty(t, niceTicks(d, d, 6))

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
		assert.Equal(t, c.want, got)

	}
}

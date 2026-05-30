# ascii-timeline

Render a horizontal, annotated timeline as styled terminal text from a small
JSON document.

The axis runs left (start) to right (end). Events are drawn beneath it in
chronological order: **point** events as a single marker (`●`), **duration**
events as a bar (`█`) spanning their range. Output uses Unicode box-drawing
characters and ANSI color by default. Everything is optional.

```text
# Project Apollo
High-level delivery plan. Point markers are milestones; bars are work phases.

Dec 30, 2025  →  Dec 30, 2026   (12 months)
Jan 2026    Mar 2026    May today    Jul 2026    Sep 2026    Nov 2026
├───────────┴───────────┴────▼──────┴───────────┴───────────┴──────────┤
 ● Kickoff (Jan 4, 2026)
      ██████████ Design phase│(Jan 29, 2026 – Mar 15, 2026) — research and specs
               ● Design review (Mar 18, 2026)
                ██████████████████████████ Build phase (Mar 20, 2026 – Jul 28, 2026) — core implementation
                             │                 ● Beta release (Aug 27, 2026)
                             │                 ████████████ Hardening (Aug 27, 2026 – Oct 26, 2026)
                             │                                   ● GA launch (Nov 30, 2026) — general availability
```

In a real terminal: the header is bold; the caption and axis are dim; each
event uses its palette color; the current date appears as a bold bright-yellow
`▼` on the axis, `today` label on the row above, and a vertical `│` line running
through the event rows below — making it easy to see what's past, ongoing, and
upcoming at a glance. The `example` subcommand generates dates relative to the
current date so today always falls mid-timeline.

## Build

This project uses [`go-toolchain`](https://github.com/wow-look-at-my/go-toolchain).
Run it (no arguments) in the repo root to tidy, test, and build:

```sh
go-toolchain
```

The binary is written to `build/ascii-timeline`.

## CLI usage

Input is a JSON document read from a file, a positional argument, or stdin:

```sh
ascii-timeline -f roadmap.json      # from a file
ascii-timeline roadmap.json         # positional file
cat roadmap.json | ascii-timeline   # from stdin

ascii-timeline example              # print a complete sample document
ascii-timeline example | ascii-timeline   # render the sample
```

Flags:

| Flag             | Description                                          |
| ---------------- | ---------------------------------------------------- |
| `-f, --file`     | Input JSON file (default: stdin)                     |
| `-w, --width`    | Axis width in columns (default 72)                   |
| `--no-color`     | Disable ANSI color (also honored via `NO_COLOR` env) |

## JSON format

```json
{
  "header":      "# Timeline",
  "description": "Free text shown under the header.",
  "start":       "2024-01-01",
  "end":         "2024-12-31",
  "width":       72,
  "events": [
    { "label": "Kickoff", "date":  "2024-01-08", "color": "green" },
    { "label": "Build",   "start": "2024-02-01", "end": "2024-06-30",
      "color": "blue", "description": "core implementation" }
  ]
}
```

| Field         | Where    | Required | Notes                                                                   |
| ------------- | -------- | -------- | ----------------------------------------------------------------------- |
| `header`      | document | no       | Rendered bold. Omit for the default `# Timeline`; set `""` for none.    |
| `description` | document | no       | Free text, word-wrapped under the header.                               |
| `start`       | document | no       | Left edge. Derived from the events (and padded) when omitted.           |
| `end`         | document | no       | Right edge. Derived from the events (and padded) when omitted.          |
| `width`       | document | no       | Axis width in columns (default 72). `-w` overrides it.                  |
| `events`      | document | no       | The annotations (see below).                                            |
| `label`       | event    | no       | Text shown next to the marker/bar.                                      |
| `description` | event    | no       | Extra dim text appended after the date.                                 |
| `date`        | event    | one of\* | A **point** event at this instant.                                      |
| `start`+`end` | event    | one of\* | A **duration** event spanning the range.                                |
| `color`       | event    | no       | Style spec (see below). Auto-assigned from a palette when omitted.      |

\* Each event needs either `date` (point) **or** both `start` and `end`
(duration).

### Dates

Many formats are accepted; the first that parses wins:

```
2006-01-02   2006/01/02   01/02/2006   2006-01   2006
"Jan 2, 2006"   "January 2, 2006"   "2 Jan 2006"   "Jan 2006"
2006-01-02 15:04   2006-01-02T15:04:05   RFC3339 (with timezone)
```

Date-only values are interpreted at midnight UTC.

### Colors

`color` accepts names and styles, optionally combined with spaces, commas or
`+`:

- Colors: `black red green yellow blue magenta cyan white gray`
- Bright: `brightred brightgreen brightyellow brightblue brightmagenta brightcyan brightwhite`
- Styles: `bold dim italic underline`
- Combined: `"bold cyan"`, `"underline red"`, `"dim"`

Unknown tokens are ignored, so a bad spec degrades to plain text rather than
failing.

## Using it from an LLM

The JSON schema is intentionally small, lenient, and self-describing so a model
can emit it directly:

- Every field is optional; the smallest valid document is a single event.
- Unknown keys are tolerated (they're ignored), so extra annotations won't break parsing.
- `ascii-timeline example` prints a complete, working document to copy from.
- `ascii-timeline --help` documents the full shape.

## Library usage

The renderer is also a Go package:

```go
package main

import (
	"os"
	"time"

	"github.com/wow-look-at-my/ascii-timeline/timeline"
)

func main() {
	// From JSON:
	tl, err := timeline.Parse(os.Stdin)
	if err != nil {
		panic(err)
	}
	_ = tl.Render(os.Stdout)

	// Or build it directly (New seeds the default header/width):
	tl = timeline.New()
	tl.Header = "# Release plan"
	tl.Events = []timeline.Event{
		{Label: "Kickoff", Date: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), Color: "green"},
		{Label: "Build",
			Start: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC),
			Color: "blue"},
	}
	_ = tl.Render(os.Stdout)
}
```

Set `Timeline.NoColor = true` to emit plain text.

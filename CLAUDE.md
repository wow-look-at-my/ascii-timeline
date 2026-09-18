# CLAUDE.md

Project notes for working in this repository.

## What this is

`ascii-timeline` renders a horizontal, annotated timeline as styled terminal
text from a JSON document. It is both a CLI and a reusable Go library.

## Layout

```
main.go              # entry point; calls cmd.Execute()
cmd/                 # cobra CLI (one command per file, self-registering)
  root.go            # root command: reads JSON (file/arg/stdin) and renders
  example.go         # `example` subcommand; generates dynamic sample JSON
  version.go         # `version` subcommand
timeline/            # rendering library (no CLI deps)
  timeline.go        # Timeline/Event types, New, Render, bounds, layout, drawToday
  parse.go           # JSON + flexible date parsing (Parse/ParseBytes)
  ticks.go           # axis tick selection and date label formats
  canvas.go          # styled rune grid with ANSI run-length output
  color.go           # color/style name -> ANSI SGR mapping
  *_test.go          # tests
```

## Build / test

Use `go-toolchain` (no arguments) in the repo root — it tidies, vets, tests
with coverage, and builds to `build/ascii-timeline`. Do not invoke bare `go`
commands.

```sh
go-toolchain
./build/ascii-timeline example | ./build/ascii-timeline   # quick visual check
```

Tests use `github.com/stretchr/testify` (`assert`/`require`). The
`github.com/wow-look-at-my/testify` fork this repo once imported has been
deleted, and go-toolchain's vet now migrates any import of it back to
upstream.

## Design notes

- The axis maps time to columns linearly: `start` -> col 0, `end` -> col
  `width-1`. Bounds are derived from events (and padded ~4%) when not given
  explicitly; an explicit `start`/`end` is used exactly so "start on the left"
  holds.
- Events are sorted chronologically and drawn one per row beneath the axis —
  simple and robust, no overlap packing.
- Axis tick labels are coarse gridlines (`tickFormat`); per-event dates are one
  notch finer (`eventDateFormat`) so events keep their day. The exact
  start/end always appear in the dim caption above the axis.
- Color: `header` is bold; caption, axis and dates are dim; markers/bars use
  the event color (auto-assigned from a palette when unset). `NoColor` (or
  `--no-color` / `NO_COLOR`) drops all escapes.
- Everything is optional; `header` defaults to `# Timeline` (a `*string` in the
  JSON layer distinguishes "omitted" from an explicit empty `""`).
- Today marker: `drawToday` runs after `drawEvents`; it places a bold-brightyellow
  `▼` on the axis row and "today" on the label row, then draws `│` down through
  all blank cells in the today column. Only fires when today is within `[rc.s, rc.e]`.

## Docs

Keep `README.md` (JSON schema, flags, sample output) and this file in sync with
any change to the JSON format, flags, or layout.

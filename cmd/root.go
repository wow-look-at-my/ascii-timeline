// Package cmd implements the ascii-timeline command-line interface.
package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/wow-look-at-my/ascii-timeline/timeline"
)

var (
	flagFile    string
	flagWidth   int
	flagNoColor bool
)

var rootCmd = &cobra.Command{
	Use:   "ascii-timeline [file]",
	Short: "Render a horizontal, annotated timeline from JSON",
	Long: `ascii-timeline renders a horizontal timeline as styled terminal text.

It reads a JSON document from a file, a positional argument, or stdin and
draws a date axis (start on the left, end on the right) annotated with events.
Each event is either a point (a single "date") or a duration (a "start" and
"end"). Header, description, bounds, width and per-event color are all optional.

JSON shape:

  {
    "header":      "# Timeline",                  // optional, bold; default "# Timeline"
    "description": "Free text shown under it.",    // optional
    "start":       "2024-01-01",                   // optional, derived from events
    "end":         "2024-12-31",                   // optional, derived from events
    "width":       72,                              // optional axis width in columns
    "events": [
      { "label": "Kickoff",  "date":  "2024-01-08", "color": "green" },
      { "label": "Build",    "start": "2024-02-01", "end": "2024-06-30", "color": "blue",
        "description": "core implementation" }
    ]
  }

Dates accept many formats (2006-01-02, "Jan 2, 2006", 2006-01, 2006, RFC3339).
Colors accept names and styles, optionally combined: "red", "cyan",
"brightgreen", "bold magenta", "dim".

Run 'ascii-timeline example' to print a complete sample document, or pipe one
straight in:  ascii-timeline example | ascii-timeline`,
	Args:          cobra.MaximumNArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		path := flagFile
		if path == "" && len(args) == 1 {
			path = args[0]
		}

		var (
			data []byte
			err  error
		)
		if path == "" || path == "-" {
			data, err = io.ReadAll(cmd.InOrStdin())
		} else {
			data, err = os.ReadFile(path)
		}
		if err != nil {
			return err
		}

		tl, err := timeline.ParseBytes(data)
		if err != nil {
			return err
		}
		if flagWidth > 0 {
			tl.Width = flagWidth
		}
		if flagNoColor || os.Getenv("NO_COLOR") != "" {
			tl.NoColor = true
		}
		return tl.Render(cmd.OutOrStdout())
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func init() {
	f := rootCmd.Flags()
	f.StringVarP(&flagFile, "file", "f", "", "input JSON file (default: stdin)")
	f.IntVarP(&flagWidth, "width", "w", 0, "axis width in columns (default 72)")
	f.BoolVar(&flagNoColor, "no-color", false, "disable ANSI color output (also honors NO_COLOR)")
}

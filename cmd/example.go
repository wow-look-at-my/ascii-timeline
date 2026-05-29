package cmd

import (
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"
)

var exampleCmd = &cobra.Command{
	Use:   "example",
	Short: "Print a complete sample timeline JSON document",
	Long: `Print a complete sample timeline document to stdout.

Save it to a file or render it directly:

  ascii-timeline example > roadmap.json
  ascii-timeline example | ascii-timeline`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		_, err := io.WriteString(cmd.OutOrStdout(), buildExampleJSON())
		return err
	},
}

// buildExampleJSON returns a sample document whose dates are computed relative
// to today so that the current date always falls mid-timeline (inside the
// ongoing "Build phase") when rendered.
func buildExampleJSON() string {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	d := func(offset int) string {
		return today.AddDate(0, 0, offset).Format("2006-01-02")
	}
	return fmt.Sprintf(`{
  "header": "# Project Apollo",
  "description": "High-level delivery plan. Point markers are milestones; bars are work phases.",
  "start": "%s",
  "end": "%s",
  "events": [
    { "label": "Kickoff",       "date":  "%s",                 "color": "green" },
    { "label": "Design phase",  "start": "%s", "end": "%s",    "color": "cyan",        "description": "research and specs" },
    { "label": "Design review", "date":  "%s",                 "color": "yellow" },
    { "label": "Build phase",   "start": "%s", "end": "%s",    "color": "blue",        "description": "core implementation" },
    { "label": "Beta release",  "date":  "%s",                 "color": "magenta" },
    { "label": "Hardening",     "start": "%s", "end": "%s",    "color": "red" },
    { "label": "GA launch",     "date":  "%s",                 "color": "brightgreen", "description": "general availability" }
  ]
}
`,
		d(-150), d(215),
		d(-145),
		d(-120), d(-75),
		d(-72),
		d(-70), d(60),
		d(90),
		d(90), d(150),
		d(185),
	)
}

func init() {
	rootCmd.AddCommand(exampleCmd)
}

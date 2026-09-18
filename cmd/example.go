package cmd

import (
	"encoding/json"
	"io"
	"time"

	"github.com/spf13/cobra"
)

// exampleEvent and exampleDoc mirror the document the timeline package reads.
// They exist so the sample is marshalled rather than formatted: a quote or a
// newline in any value would otherwise break the JSON.
type exampleEvent struct {
	Label       string `json:"label"`
	Date        string `json:"date,omitempty"`
	Start       string `json:"start,omitempty"`
	End         string `json:"end,omitempty"`
	Color       string `json:"color"`
	Description string `json:"description,omitempty"`
}

type exampleDoc struct {
	Header      string         `json:"header"`
	Description string         `json:"description"`
	Start       string         `json:"start"`
	End         string         `json:"end"`
	Events      []exampleEvent `json:"events"`
}

var exampleCmd = &cobra.Command{
	Use:   "example",
	Short: "Print a complete sample timeline JSON document",
	Long: `Print a complete sample timeline document to stdout.

Save it to a file or render it directly:

  ascii-timeline example > roadmap.json
  ascii-timeline example | ascii-timeline`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		doc, err := buildExampleJSON()
		if err != nil {
			return err
		}
		_, err = io.WriteString(cmd.OutOrStdout(), doc)
		return err
	},
}

// buildExampleJSON returns a sample document whose dates are computed relative
// to today so that the current date always falls mid-timeline (inside the
// ongoing "Build phase") when rendered.
func buildExampleJSON() (string, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	d := func(offset int) string {
		return today.AddDate(0, 0, offset).Format("2006-01-02")
	}

	doc := exampleDoc{
		Header:      "# Project Apollo",
		Description: "High-level delivery plan. Point markers are milestones, and bars are work phases.",
		Start:       d(-150),
		End:         d(215),
		Events: []exampleEvent{
			{Label: "Kickoff", Date: d(-145), Color: "green"},
			{Label: "Design phase", Start: d(-120), End: d(-75), Color: "cyan", Description: "research and specs"},
			{Label: "Design review", Date: d(-72), Color: "yellow"},
			{Label: "Build phase", Start: d(-70), End: d(60), Color: "blue", Description: "core implementation"},
			{Label: "Beta release", Date: d(90), Color: "magenta"},
			{Label: "Hardening", Start: d(90), End: d(150), Color: "red"},
			{Label: "GA launch", Date: d(185), Color: "brightgreen", Description: "general availability"},
		},
	}

	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", err
	}
	return string(out) + "\n", nil
}

func init() {
	rootCmd.AddCommand(exampleCmd)
}

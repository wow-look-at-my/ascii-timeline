package cmd

import (
	_ "embed"
	"io"

	"github.com/spf13/cobra"
)

//go:embed example.json
var exampleJSON string

var exampleCmd = &cobra.Command{
	Use:   "example",
	Short: "Print a complete sample timeline JSON document",
	Long: `Print a complete sample timeline document to stdout.

Save it to a file or render it directly:

  ascii-timeline example > roadmap.json
  ascii-timeline example | ascii-timeline`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		_, err := io.WriteString(cmd.OutOrStdout(), exampleJSON)
		return err
	},
}

func init() {
	rootCmd.AddCommand(exampleCmd)
}

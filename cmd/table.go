package cmd

import "github.com/spf13/cobra"

var tableCmd = &cobra.Command{
	Use:   "table [file]",
	Short: "Interactive data table",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runChart("table", args)
	},
}

func init() {
	rootCmd.AddCommand(tableCmd)
}

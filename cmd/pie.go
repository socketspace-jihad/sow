package cmd

import "github.com/spf13/cobra"

var pieCmd = &cobra.Command{
	Use:   "pie [file]",
	Short: "Pie chart (proportional colored bar with legend)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runChart("pie", args)
	},
}

func init() {
	rootCmd.AddCommand(pieCmd)
}

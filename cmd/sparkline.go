package cmd

import "github.com/spf13/cobra"

var sparklineCmd = &cobra.Command{
	Use:   "sparkline [file]",
	Short: "Compact sparkline chart",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runChart("sparkline", args)
	},
}

func init() {
	rootCmd.AddCommand(sparklineCmd)
}

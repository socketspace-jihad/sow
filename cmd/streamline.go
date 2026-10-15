package cmd

import "github.com/spf13/cobra"

var streamlineCmd = &cobra.Command{
	Use:   "streamline [file]",
	Short: "Scrolling streamline chart (ideal for live data)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runChart("streamline", args)
	},
}

func init() {
	rootCmd.AddCommand(streamlineCmd)
}

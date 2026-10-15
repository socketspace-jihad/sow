package cmd

import "github.com/spf13/cobra"

var lineCmd = &cobra.Command{
	Use:   "line [file]",
	Short: "Line chart",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runChart("line", args)
	},
}

func init() {
	rootCmd.AddCommand(lineCmd)
}

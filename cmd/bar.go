package cmd

import "github.com/spf13/cobra"

var barCmd = &cobra.Command{
	Use:   "bar [file]",
	Short: "Vertical bar chart",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runChart("bar", args)
	},
}

var hbarCmd = &cobra.Command{
	Use:   "hbar [file]",
	Short: "Horizontal bar chart",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runChart("hbar", args)
	},
}

func init() {
	rootCmd.AddCommand(barCmd)
	rootCmd.AddCommand(hbarCmd)
}

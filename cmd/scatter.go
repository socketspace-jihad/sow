package cmd

import "github.com/spf13/cobra"

var scatterCmd = &cobra.Command{
	Use:   "scatter [file]",
	Short: "Scatter / dot plot",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runChart("scatter", args)
	},
}

var dotCmd = &cobra.Command{
	Use:   "dot [file]",
	Short: "Dot plot (alias for scatter)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runChart("dot", args)
	},
}

func init() {
	rootCmd.AddCommand(scatterCmd)
	rootCmd.AddCommand(dotCmd)
}

package cmd

import "github.com/spf13/cobra"

var heatmapCmd = &cobra.Command{
	Use:   "heatmap [file]",
	Short: "Heatmap visualization",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runChart("heatmap", args)
	},
}

func init() {
	rootCmd.AddCommand(heatmapCmd)
}

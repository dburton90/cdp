package cmd

import (
	"github.com/spf13/cobra"
)

var refreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Rescan directories and rebuild cache",
	Run: func(cmd *cobra.Command, args []string) {
		PerformRefresh()
	},
}

func init() {
	rootCmd.AddCommand(refreshCmd)
}

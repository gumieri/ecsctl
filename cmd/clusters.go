package cmd

import (
	"github.com/spf13/cobra"
)

func clustersRun(cmd *cobra.Command, args []string) {
	cmd.Help()
}

var clustersCmd = &cobra.Command{
	Use:   "clusters [command]",
	Short: "commands to manage clusters",
	Run:   clustersRun,
}

func init() {
	rootCmd.AddCommand(clustersCmd)
}

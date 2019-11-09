package cmd

import (
	"github.com/spf13/cobra"
)

func instancesRun(cmd *cobra.Command, args []string) {
	cmd.Help()
}

var instancesCmd = &cobra.Command{
	Use:     "instances [command]",
	Short:   "Commands to manage ECS instances",
	Aliases: []string{"instance", "i"},
	Run:     instancesRun,
}

func init() {
	rootCmd.AddCommand(instancesCmd)
}

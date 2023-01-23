package cmd

import (
	"github.com/spf13/cobra"
)

func scheduledTasksRun(cmd *cobra.Command, args []string) {
	cmd.Help()
}

var scheduledTasksCmd = &cobra.Command{
	Use:     "scheduled-tasks [command]",
	Short:   "Commands to manage scheduled tasks",
	Aliases: []string{"schedule", "st"},
	Run:     scheduledTasksRun,
}

func init() {
	rootCmd.AddCommand(scheduledTasksCmd)
}

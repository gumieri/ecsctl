package cmd

import (
	"github.com/spf13/cobra"
)

func taskDefinitionsRun(cmd *cobra.Command, args []string) {
	cmd.Help()
}

var taskDefinitionsCmd = &cobra.Command{
	Use:     "task-definitions [command]",
	Short:   "commands to manage Task Definitions",
	Aliases: []string{"task-definition", "tasks", "task", "t"},
	Run:     taskDefinitionsRun,
}

func init() {
	rootCmd.AddCommand(taskDefinitionsCmd)
}

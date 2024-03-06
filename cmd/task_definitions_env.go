package cmd

import (
	"github.com/spf13/cobra"
)

func taskDefinitionsEnvRun(cmd *cobra.Command, args []string) {
	cmd.Help()
}

var taskDefinitionsEnvCmd = &cobra.Command{
	Use:   "env [command]",
	Short: "Commands to manage Task Definitions' environment variables",
	Run:   taskDefinitionsEnvRun,
}

func init() {
	taskDefinitionsCmd.AddCommand(taskDefinitionsEnvCmd)
}

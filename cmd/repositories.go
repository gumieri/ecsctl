package cmd

import (
	"github.com/spf13/cobra"
)

func repositoriesRun(cmd *cobra.Command, args []string) {
	cmd.Help()
}

var repositoriesCmd = &cobra.Command{
	Use:     "repositories [command]",
	Short:   "Commands to manage repositories (ECR)",
	Aliases: []string{"repository", "ecr", "r"},
	Run:     repositoriesRun,
}

func init() {
	rootCmd.AddCommand(repositoriesCmd)
}

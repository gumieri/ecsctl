package cmd

import (
	"github.com/spf13/cobra"
)

func servicesRun(cmd *cobra.Command, args []string) {
	cmd.Help()
}

var servicesCmd = &cobra.Command{
	Use:     "services [command]",
	Short:   "commands to manage services",
	Aliases: []string{"service", "s"},
	Run:     servicesRun,
}

func init() {
	rootCmd.AddCommand(servicesCmd)
}

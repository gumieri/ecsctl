package cmd

import (
	"github.com/spf13/cobra"
)

func servicesRun(cmd *cobra.Command, args []string) {
}

var servicesCmd = &cobra.Command{
	Use:   "services",
	Short: "commands to manage services",
	Run:   servicesRun,
}

func init() {
	rootCmd.AddCommand(servicesCmd)
}

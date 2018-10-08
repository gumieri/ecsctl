package cmd

import (
	"github.com/spf13/cobra"
)

// VERSION of the ecsctl
var VERSION = "v0.3.7"

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of ecsctl",
	Run: func(cmd *cobra.Command, args []string) {
		typist.Println(VERSION)
	},
}

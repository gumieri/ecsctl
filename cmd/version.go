package cmd

import (
	"github.com/spf13/cobra"
)

// VERSION of the ecsctl
var VERSION string

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

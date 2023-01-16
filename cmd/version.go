package cmd

import (
	"github.com/spf13/cobra"
)

// Version of the ecsctl
var Version string

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of ecsctl",
	Run:   func(cmd *cobra.Command, args []string) { t.Outln(Version) },
}

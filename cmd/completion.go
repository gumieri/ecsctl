package cmd

import (
	"errors"
	"os"

	"github.com/spf13/cobra"
)

func completionRun(cmd *cobra.Command, args []string) {
	switch args[0] {
	case "bash":
		rootCmd.GenBashCompletion(os.Stdout)
	case "zsh":
		rootCmd.GenZshCompletion(os.Stdout)
	default:
		t.Must(errors.New("specied shell is not yet supported"))
	}
}

var completionCmd = &cobra.Command{
	Use:   "completion [shell]",
	Short: "Output the completion script for the specified shell language ('bash' or 'zsh')",
	Args:  cobra.ExactArgs(1),
	Run:   completionRun,
}

func init() {
	rootCmd.AddCommand(completionCmd)
}

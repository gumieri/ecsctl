package cmd

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/spf13/cobra"
)

func repositoriesCreateRun(cmd *cobra.Command, repositories []string) {
	for _, repository := range repositories {
		t.Must(ecrI.CreateRepository(&ecr.CreateRepositoryInput{
			RepositoryName: aws.String(repository),
		}))
	}
}

var repositoriesCreateCmd = &cobra.Command{
	Use:   "create [repositories...]",
	Short: "Create repositories",
	Args:  cobra.MinimumNArgs(1),
	Run:   repositoriesCreateRun,
}

func init() {
	repositoriesCmd.AddCommand(repositoriesCreateCmd)
}

package cmd

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/spf13/cobra"
)

func repositoriesCreateRun(cmd *cobra.Command, repositories []string) {
	for _, repository := range repositories {
		_, err := ecrI.CreateRepository(&ecr.CreateRepositoryInput{
			RepositoryName: aws.String(repository),
		})

		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
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

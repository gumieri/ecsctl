package cmd

import (
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/spf13/cobra"
)

func repositoriesDeleteRun(cmd *cobra.Command, repositories []string) {
	repositoriesDescription, err := ecrI.DescribeRepositories(&ecr.DescribeRepositoriesInput{
		RepositoryNames: aws.StringSlice(repositories),
	})

	// Ignore only ErrCodeRepositoryNotFoundException
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() != ecr.ErrCodeRepositoryNotFoundException {
				t.Must(err)
			}
		} else {
			t.Must(err)
		}
	}

	var foundRepositories []*ecr.Repository
	var missing []string
	for _, named := range repositories {
		var found bool

		for _, repository := range repositoriesDescription.Repositories {
			if aws.StringValue(repository.RepositoryName) == named {
				foundRepositories = append(foundRepositories, repository)
				found = true
				break
			}
		}

		if !found {
			missing = append(missing, named)
		}
	}

	if !force && len(missing) > 0 {
		t.Must(errors.New("Some repositories were not found:\n\t" + strings.Join(missing, "\n\t")))
	}

	if !force && !yes && len(foundRepositories) > 0 {
		t.Infoln("repositories to be deleted:")
		for _, repository := range foundRepositories {
			t.Infoln(aws.StringValue(repository.RepositoryArn))
		}

		if !t.Confirm("Do you really want to delete these repositories?") {
			return
		}
	}

	for _, repository := range foundRepositories {
		_, err := ecrI.DeleteRepository(&ecr.DeleteRepositoryInput{
			RepositoryName: repository.RepositoryName,
		})

		t.Must(err)

		t.Infof("%s deleted\n", aws.StringValue(repository.RepositoryArn))
	}
}

var repositoriesDeleteCmd = &cobra.Command{
	Use:   "delete [repositories...]",
	Short: "Delete repositories",
	Args:  cobra.MinimumNArgs(1),
	Run:   repositoriesDeleteRun,
}

func init() {
	repositoriesCmd.AddCommand(repositoriesDeleteCmd)
	flags := repositoriesDeleteCmd.Flags()
	flags.BoolVarP(&yes, "yes", "y", false, yesSpec)
	flags.BoolVarP(&force, "force", "f", false, forceSpec)
}

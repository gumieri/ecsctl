package cmd

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/spf13/cobra"
)

func taskDefinitionsListRun(cmd *cobra.Command, args []string) {
	input := &ecs.ListTaskDefinitionFamiliesInput{
		Status: aws.String(strings.ToUpper(status)),
	}

	if len(args) > 0 {
		input.FamilyPrefix = aws.String(args[0])
	}

	var nextToken *string
	for {
		if nextToken != nil {
			input.NextToken = nextToken
		}

		result, err := ecsI.ListTaskDefinitionFamilies(input)
		t.Must(err)

		for _, f := range result.Families {
			t.Outln(aws.StringValue(f))
		}

		if result.NextToken == nil {
			break
		}

		nextToken = result.NextToken
	}
}

var taskDefinitionsListCmd = &cobra.Command{
	Use:              "list [prefix filter]",
	Short:            "List all Task Definition Families",
	Aliases:          []string{"l"},
	Args:             cobra.MaximumNArgs(1),
	Run:              taskDefinitionsListRun,
	PersistentPreRun: persistentPreRun,
}

func init() {
	taskDefinitionsCmd.AddCommand(taskDefinitionsListCmd)

	flags := taskDefinitionsListCmd.Flags()

	flags.StringVarP(&status, "status", "s", "active", statusSpec)
}

package cmd

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/spf13/cobra"
)

func taskDefinitionsListRun(cmd *cobra.Command, args []string) {
	input := &ecs.ListTaskDefinitionFamiliesInput{}

	if len(args) > 0 {
		input.FamilyPrefix = aws.String(args[0])
	}

	var nextToken *string
	for {
		if nextToken != nil {
			input.NextToken = nextToken
		}

		result, err := ecsI.ListTaskDefinitionFamilies(input)

		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		for _, f := range result.Families {
			fmt.Println(aws.StringValue(f))
		}

		if result.NextToken == nil {
			break
		}

		nextToken = result.NextToken
	}
}

var taskDefinitionsListCmd = &cobra.Command{
	Use:   "list [prefix filter]",
	Short: "List all Task Definition Families",
	Args:  cobra.MaximumNArgs(1),
	Run:   taskDefinitionsListRun,
}

func init() {
	taskDefinitionsCmd.AddCommand(taskDefinitionsListCmd)
}

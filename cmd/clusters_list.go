package cmd

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/spf13/cobra"
)

func clustersListRun(cmd *cobra.Command, clusters []string) {
	input := &ecs.ListClustersInput{}

	var nextToken *string
	for {
		if nextToken != nil {
			input.NextToken = nextToken
		}

		result, err := ecsI.ListClusters(input)
		t.Must(err)

		for _, clusterARN := range result.ClusterArns {
			if listARN {
				t.Outln(aws.StringValue(clusterARN))
				continue
			}

			parsedARN, err := arn.Parse(aws.StringValue(clusterARN))
			t.Must(err)
			parsedResource := strings.Split(parsedARN.Resource, "/")
			t.Outln(parsedResource[len(parsedResource)-1])
		}

		if result.NextToken == nil {
			break
		}

		nextToken = result.NextToken
	}
}

var clustersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List clusters",
	Run:   clustersListRun,
}

func init() {
	clustersCmd.AddCommand(clustersListCmd)

	flags := clustersListCmd.Flags()

	flags.BoolVar(&listARN, "arn", false, listARNSpec)
}

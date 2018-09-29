package cmd

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/spf13/cobra"
)

func clustersCreateRun(cmd *cobra.Command, clusters []string) {
	for _, cluster := range clusters {
		_, err := ecsI.CreateCluster(&ecs.CreateClusterInput{
			ClusterName: aws.String(cluster),
		})

		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}
}

var clustersCreateCmd = &cobra.Command{
	Use: "create [clusters...]",
	Short: `Create empty clusters.
If you do not specify a name for your cluster, you create a cluster named default.
Up to 255 letters (uppercase and lowercase), numbers, hyphens, and underscores are allowed.`,
	Run: clustersCreateRun,
}

func init() {
	clustersCmd.AddCommand(clustersCreateCmd)
}

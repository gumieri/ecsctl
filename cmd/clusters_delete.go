package cmd

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/spf13/cobra"
)

func clustersDeleteRun(cmd *cobra.Command, clusters []string) {
	clustersDescription, err := ecsI.DescribeClusters(&ecs.DescribeClustersInput{
		Clusters: aws.StringSlice(clusters),
	})
	must(err)

	failures := clustersDescription.Failures
	if len(failures) > 0 && !force {
		fmt.Println("Some clusters were not found:")
		for _, notFound := range failures {
			fmt.Println(aws.StringValue(notFound.Arn))
		}
		os.Exit(1)
	}

	foundClusters := clustersDescription.Clusters
	if len(foundClusters) > 0 && !yes {
		fmt.Println("Do you want to delete the clusters:")
		for _, foundCluster := range foundClusters {
			fmt.Println(aws.StringValue(foundCluster.ClusterArn))
		}
	}

	// for _, cluster := range clusters {
	// 	_, err := ecsI.DeleteCluster(&ecs.DeleteClusterInput{
	// 		Cluster: aws.String(cluster),
	// 	})

	// 	if err != nil {
	// 		fmt.Println(err.Error())
	// 		os.Exit(1)
	// 	}
	// }
}

var clustersDeleteCmd = &cobra.Command{
	Use:   "delete [clusters...]",
	Short: "Delete empty clusters",
	Args:  cobra.MinimumNArgs(1),
	Run:   clustersDeleteRun,
}

func init() {
	clustersCmd.AddCommand(clustersDeleteCmd)
	flags := clustersDeleteCmd.Flags()
	flags.BoolVarP(&yes, "yes", "y", false, yesSpec)
	flags.BoolVarP(&force, "force", "f", false, forceSpec)
}

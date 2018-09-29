package cmd

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/gumieri/cli/confirm"
	"github.com/spf13/cobra"
)

func clustersDeleteRun(cmd *cobra.Command, clusters []string) {
	clustersDescription, err := ecsI.DescribeClusters(&ecs.DescribeClustersInput{
		Clusters: aws.StringSlice(clusters),
	})
	must(err)

	failures := clustersDescription.Failures
	if !force && len(failures) > 0 {
		fmt.Println("Some clusters were not found:")
		for _, notFound := range failures {
			fmt.Println(aws.StringValue(notFound.Arn))
		}
		os.Exit(1)
	}

	foundClusters := clustersDescription.Clusters
	var activeClusters []*ecs.Cluster
	for _, cluster := range foundClusters {
		if aws.StringValue(cluster.Status) == "ACTIVE" {
			activeClusters = append(activeClusters, cluster)
		}
	}

	if !force && !yes && len(activeClusters) > 0 {
		fmt.Println("clusters to be deleted:")
		for _, cluster := range activeClusters {
			fmt.Println(aws.StringValue(cluster.ClusterArn))
		}

		if !confirm.Confirm("Do you really want to delete these clusters?") {
			return
		}
	}

	for _, cluster := range activeClusters {
		_, err := ecsI.DeleteCluster(&ecs.DeleteClusterInput{
			Cluster: cluster.ClusterArn,
		})

		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		fmt.Printf("%s deleted\n", aws.StringValue(cluster.ClusterArn))
	}
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

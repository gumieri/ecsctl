package cmd

import (
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/spf13/cobra"
)

func clustersDeleteRun(cmd *cobra.Command, clusters []string) {
	clustersDescription, err := ecsI.DescribeClusters(&ecs.DescribeClustersInput{
		Clusters: aws.StringSlice(clusters),
	})
	t.Must(err)

	var missing []string
	var activeClusters []*ecs.Cluster

	foundClusters := clustersDescription.Clusters
	for _, cluster := range foundClusters {
		if aws.StringValue(cluster.Status) == "ACTIVE" {
			activeClusters = append(activeClusters, cluster)
		} else {
			missing = append(missing, aws.StringValue(cluster.ClusterArn))
		}
	}

	for _, notFound := range clustersDescription.Failures {
		missing = append(missing, aws.StringValue(notFound.Arn))
	}

	if !force && len(missing) > 0 {
		t.Must(errors.New("Some clusters were not found:\n\t" + strings.Join(missing, "\n\t")))
	}

	if !force && !yes && len(activeClusters) > 0 {
		t.Infoln("clusters to be deleted:")
		for _, cluster := range activeClusters {
			t.Infoln(aws.StringValue(cluster.ClusterArn))
		}

		if !t.Confirm("Do you really want to delete these clusters?") {
			return
		}
	}

	for _, cluster := range activeClusters {
		_, err := ecsI.DeleteCluster(&ecs.DeleteClusterInput{
			Cluster: cluster.ClusterArn,
		})

		t.Must(err)

		t.Infof("%s deleted\n", aws.StringValue(cluster.ClusterArn))
	}
}

var clustersDeleteCmd = &cobra.Command{
	Use:              "delete [clusters...]",
	Short:            "Delete clusters",
	Args:             cobra.MinimumNArgs(1),
	Run:              clustersDeleteRun,
	PersistentPreRun: persistentPreRun,
}

func init() {
	clustersCmd.AddCommand(clustersDeleteCmd)
	flags := clustersDeleteCmd.Flags()
	flags.BoolVarP(&yes, "yes", "y", false, yesSpec)
	flags.BoolVarP(&force, "force", "f", false, forceSpec)
}

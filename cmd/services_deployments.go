package cmd

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func servicesDeploymentsRun(cmd *cobra.Command, args []string) {
	var services []*string
	for _, s := range args {
		services = append(services, aws.String(s))
	}

	cluster := viper.GetString("cluster")

	clustersDescription, err := ecsI.DescribeClusters(&ecs.DescribeClustersInput{
		Clusters: []*string{aws.String(cluster)},
	})

	t.Must(err)

	if len(clustersDescription.Clusters) == 0 {
		t.Exitln("Source Cluster informed not found")
	}

	c := clustersDescription.Clusters[0]

	for {
		servicesDescription, err := ecsI.DescribeServices(&ecs.DescribeServicesInput{
			Cluster:  c.ClusterName,
			Services: services,
		})

		t.Must(err)

		for _, d := range servicesDescription.Services[0].Deployments {
			if aws.StringValue(d.Status) == "PRIMARY" {
				switch aws.StringValue(d.RolloutState) {
				case ecs.DeploymentRolloutStateCompleted:
					t.Outln("Complete!")
					t.Exit(nil)
				case ecs.DeploymentRolloutStateFailed:
					t.Exitf("Failed! Reason: %s\n", aws.StringValue(d.RolloutStateReason))
				}
			}
		}

	}
}

var servicesDeploymentsCmd = &cobra.Command{
	Use:              "deployments [service]",
	Short:            "Query for the deployments state of specified service",
	Args:             cobra.MinimumNArgs(1),
	Run:              servicesDeploymentsRun,
	PersistentPreRun: persistentPreRun,
}

func init() {
	servicesCmd.AddCommand(servicesDeploymentsCmd)

	flags := servicesDeploymentsCmd.Flags()

	flags.StringP("cluster", "c", "", requiredSpec+clusterSpec)
	viper.BindPFlag("cluster", servicesDeploymentsCmd.Flags().Lookup("cluster"))

	servicesDeploymentsCmd.MarkFlagRequired("cluster")

}

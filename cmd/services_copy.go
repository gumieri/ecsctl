package cmd

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func servicesCopyRun(cmd *cobra.Command, services []string) {
	cluster := viper.GetString("cluster")

	targetClustersDescription, err := ecsI.DescribeClusters(&ecs.DescribeClustersInput{
		Clusters: []*string{
			aws.String(toCluster),
		},
	})

	t.Must(err)

	if len(targetClustersDescription.Clusters) == 0 {
		t.Exitf("Target Cluster informed not found")
	}

	targetC := targetClustersDescription.Clusters[0]

	clustersDescription, err := ecsI.DescribeClusters(&ecs.DescribeClustersInput{
		Clusters: []*string{aws.String(cluster)},
	})

	t.Must(err)

	if len(clustersDescription.Clusters) == 0 {
		t.Exitf("Source Cluster informed not found")
	}

	c := clustersDescription.Clusters[0]

	servicesDescription, err := ecsI.DescribeServices(&ecs.DescribeServicesInput{
		Cluster:  c.ClusterName,
		Services: aws.StringSlice(services),
	})

	if len(servicesDescription.Services) < len(services) {
		t.Exitf("One or more services informed was not found")
	}

	for _, s := range servicesDescription.Services {
		ecsI.CreateService(&ecs.CreateServiceInput{
			Cluster:                       targetC.ClusterName,
			DeploymentConfiguration:       s.DeploymentConfiguration,
			DeploymentController:          s.DeploymentController,
			DesiredCount:                  s.DesiredCount,
			EnableECSManagedTags:          s.EnableECSManagedTags,
			HealthCheckGracePeriodSeconds: s.HealthCheckGracePeriodSeconds,
			LaunchType:                    s.LaunchType,
			LoadBalancers:                 s.LoadBalancers,
			NetworkConfiguration:          s.NetworkConfiguration,
			PlacementConstraints:          s.PlacementConstraints,
			PlacementStrategy:             s.PlacementStrategy,
			PlatformVersion:               s.PlatformVersion,
			PropagateTags:                 s.PropagateTags,
			Role:                          s.RoleArn,
			SchedulingStrategy:            s.SchedulingStrategy,
			ServiceName:                   s.ServiceName,
			ServiceRegistries:             s.ServiceRegistries,
			Tags:                          s.Tags,
			TaskDefinition:                s.TaskDefinition,
		})
	}
}

var servicesCopyCmd = &cobra.Command{
	Use:   "copy [services...]",
	Short: "Copy a service to another cluster",
	Args:  cobra.MinimumNArgs(1),
	Run:   servicesCopyRun,
}

func init() {
	servicesCmd.AddCommand(servicesCopyCmd)

	flags := servicesCopyCmd.Flags()

	flags.StringVarP(&toCluster, "to-cluster", "t", "", requiredSpec+toClusterSpec)
	flags.StringP("cluster", "c", "", requiredSpec+clusterSpec)
	viper.BindPFlag("cluster", servicesCopyCmd.Flags().Lookup("cluster"))

	servicesCopyCmd.MarkFlagRequired("cluster")
	servicesCopyCmd.MarkFlagRequired("to-cluster")

}

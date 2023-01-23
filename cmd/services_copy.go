package cmd

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func servicesCopyRun(cmd *cobra.Command, services []string) {
	targetClustersDescription, err := ecsI.DescribeClusters(&ecs.DescribeClustersInput{
		Clusters: []*string{
			aws.String(toCluster),
		},
	})

	t.Must(err)

	if len(targetClustersDescription.Clusters) == 0 {
		t.Exitln("Target Cluster informed not found")
	}

	targetC := targetClustersDescription.Clusters[0]

	clustersDescription, err := ecsI.DescribeClusters(&ecs.DescribeClustersInput{
		Clusters: []*string{aws.String(cluster)},
	})

	t.Must(err)

	if len(clustersDescription.Clusters) == 0 {
		t.Exitln("Source Cluster informed not found")
	}

	c := clustersDescription.Clusters[0]

	servicesDescription, err := ecsI.DescribeServices(&ecs.DescribeServicesInput{
		Cluster:  c.ClusterName,
		Services: aws.StringSlice(services),
	})

	if len(servicesDescription.Services) < len(services) {
		t.Exitln("One or more services informed was not found")
	}

	for _, s := range servicesDescription.Services {
		input := &ecs.CreateServiceInput{
			Cluster:                       targetC.ClusterName,
			CapacityProviderStrategy:      s.CapacityProviderStrategy,
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
			Role:                          s.RoleArn,
			SchedulingStrategy:            s.SchedulingStrategy,
			ServiceName:                   s.ServiceName,
			ServiceRegistries:             s.ServiceRegistries,
			Tags:                          s.Tags,
			TaskDefinition:                s.TaskDefinition,
		}

		if aws.StringValue(s.PropagateTags) != "NONE" {
			input.PropagateTags = s.PropagateTags
		}

		if _, err := ecsI.CreateService(input); err != nil {
			t.Errorln(err)
		}
	}
}

var servicesCopyCmd = &cobra.Command{
	Use:              "copy [services...]",
	Short:            "Copy a service to another cluster",
	Args:             cobra.MinimumNArgs(1),
	Run:              servicesCopyRun,
	PersistentPreRun: persistentPreRun,
}

func init() {
	servicesCmd.AddCommand(servicesCopyCmd)

	flags := servicesCopyCmd.Flags()

	flags.StringVarP(&toCluster, "to-cluster", "t", "", requiredSpec+toClusterSpec)
	flags.StringVarP(&cluster, "cluster", "c", "", requiredSpec+clusterSpec)
	viper.BindPFlag("cluster", servicesCopyCmd.Flags().Lookup("cluster"))

	servicesCopyCmd.MarkFlagRequired("to-cluster")
}

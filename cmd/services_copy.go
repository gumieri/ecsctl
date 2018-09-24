package cmd

import (
	"errors"
	"fmt"
	"os"

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

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if len(targetClustersDescription.Clusters) == 0 {
		fmt.Println(errors.New("Target Cluster informed not found"))
		os.Exit(1)
	}

	targetC := targetClustersDescription.Clusters[0]

	clustersDescription, err := ecsI.DescribeClusters(&ecs.DescribeClustersInput{
		Clusters: []*string{
			aws.String(cluster),
		},
	})

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if len(clustersDescription.Clusters) == 0 {
		fmt.Println(errors.New("Source Cluster informed not found"))
		os.Exit(1)
	}

	c := clustersDescription.Clusters[0]

	servicesDescription, err := ecsI.DescribeServices(&ecs.DescribeServicesInput{
		Cluster:  c.ClusterName,
		Services: aws.StringSlice(services),
	})

	if len(servicesDescription.Services) < len(services) {
		fmt.Println(errors.New("One or more services informed was not found"))
		os.Exit(1)
	}

	for _, s := range servicesDescription.Services {
		ecsI.CreateService(&ecs.CreateServiceInput{
			Cluster:                       targetC.ClusterName,
			DeploymentConfiguration:       s.DeploymentConfiguration,
			DesiredCount:                  s.DesiredCount,
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
	flags.StringVarP(&cluster, "cluster", "c", "", requiredSpec+clusterSpec)

	servicesCopyCmd.MarkFlagRequired("cluster")
	servicesCopyCmd.MarkFlagRequired("to-cluster")

	viper.BindPFlag("cluster", servicesCopyCmd.Flags().Lookup("cluster"))
}

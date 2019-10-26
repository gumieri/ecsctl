package cmd

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func servicesDeployRun(cmd *cobra.Command, args []string) {
	service := args[0]

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
		fmt.Println(errors.New("Cluster informed not found"))
		os.Exit(1)
	}

	c := clustersDescription.Clusters[0]

	servicesDescription, err := ecsI.DescribeServices(&ecs.DescribeServicesInput{
		Cluster: c.ClusterName,
		Services: []*string{
			aws.String(service),
		},
	})

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if len(servicesDescription.Services) == 0 {
		fmt.Println(errors.New("Service informed not found"))
		os.Exit(1)
	}

	s := servicesDescription.Services[0]

	tdDescription, err := ecsI.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
		TaskDefinition: s.TaskDefinition,
	})

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	td := tdDescription.TaskDefinition

	var cdToUpdate *ecs.ContainerDefinition

	if containerName == "" {
		cdToUpdate = td.ContainerDefinitions[0]
	} else {
		for _, cd := range td.ContainerDefinitions {
			if aws.StringValue(cd.Name) == containerName {
				cdToUpdate = cd
				break
			}
		}
	}

	if cdToUpdate == nil {
		fmt.Println(fmt.Errorf("No container on the Task Family %s", aws.StringValue(td.Family)))
		os.Exit(1)
	}

	if tag != "" {
		image = strings.Split(aws.StringValue(cdToUpdate.Image), ":")[0] + ":" + tag
	}

	cdToUpdate.Image = aws.String(image)

	newTDDescription, err := ecsI.RegisterTaskDefinition(&ecs.RegisterTaskDefinitionInput{
		ContainerDefinitions:    td.ContainerDefinitions,
		Cpu:                     td.Cpu,
		ExecutionRoleArn:        td.ExecutionRoleArn,
		Family:                  td.Family,
		IpcMode:                 td.IpcMode,
		Memory:                  td.Memory,
		NetworkMode:             td.NetworkMode,
		PidMode:                 td.PidMode,
		PlacementConstraints:    td.PlacementConstraints,
		ProxyConfiguration:      td.ProxyConfiguration,
		RequiresCompatibilities: td.RequiresCompatibilities,
		TaskRoleArn:             td.TaskRoleArn,
		Volumes:                 td.Volumes,
	})

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	newTD := newTDDescription.TaskDefinition
	oldFamilyRevision := aws.StringValue(td.Family) + ":" + strconv.FormatInt(aws.Int64Value(td.Revision), 10)

	_, err = ecsI.DeregisterTaskDefinition(&ecs.DeregisterTaskDefinitionInput{
		TaskDefinition: aws.String(oldFamilyRevision),
	})

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	newFamilyRevision := aws.StringValue(newTD.Family) + ":" + strconv.FormatInt(aws.Int64Value(newTD.Revision), 10)

	_, err = ecsI.UpdateService(&ecs.UpdateServiceInput{
		Cluster:        c.ClusterName,
		Service:        aws.String(service),
		TaskDefinition: aws.String(newFamilyRevision),
	})

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

var servicesDeployCmd = &cobra.Command{
	Use:   "deploy [service]",
	Short: "Deploy a service",
	Args:  cobra.ExactArgs(1),
	Run:   servicesDeployRun,
}

func init() {
	servicesCmd.AddCommand(servicesDeployCmd)

	flags := servicesDeployCmd.Flags()

	flags.StringVar(&containerName, "container", "", containerNameSpec)

	flags.StringVarP(&tag, "tag", "t", "", tagSpec)
	flags.StringVarP(&image, "image", "i", "", imageSpec)
	flags.StringVarP(&cluster, "cluster", "c", "", requiredSpec+clusterSpec)

	servicesDeployCmd.MarkFlagRequired("cluster")

	viper.BindPFlag("cluster", servicesDeployCmd.Flags().Lookup("cluster"))
}

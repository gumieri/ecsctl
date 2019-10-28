package cmd

import (
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

	t.Must(err)

	if len(clustersDescription.Clusters) == 0 {
		t.Exitf("Cluster informed not found")
	}

	c := clustersDescription.Clusters[0]

	servicesDescription, err := ecsI.DescribeServices(&ecs.DescribeServicesInput{
		Cluster: c.ClusterName,
		Services: []*string{
			aws.String(service),
		},
	})

	t.Must(err)

	if len(servicesDescription.Services) == 0 {
		t.Exitf("Service informed not found")
	}

	s := servicesDescription.Services[0]

	tdDescription, err := ecsI.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
		TaskDefinition: s.TaskDefinition,
	})

	t.Must(err)

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
		t.Exitf("No container on the Task Family %s", aws.StringValue(td.Family))
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

	t.Must(err)

	newTD := newTDDescription.TaskDefinition
	oldFamilyRevision := aws.StringValue(td.Family) + ":" + strconv.FormatInt(aws.Int64Value(td.Revision), 10)

	t.Must(ecsI.DeregisterTaskDefinition(&ecs.DeregisterTaskDefinitionInput{
		TaskDefinition: aws.String(oldFamilyRevision),
	}))

	newFamilyRevision := aws.StringValue(newTD.Family) + ":" + strconv.FormatInt(aws.Int64Value(newTD.Revision), 10)

	t.Must(ecsI.UpdateService(&ecs.UpdateServiceInput{
		Cluster:        c.ClusterName,
		Service:        aws.String(service),
		TaskDefinition: aws.String(newFamilyRevision),
	}))
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

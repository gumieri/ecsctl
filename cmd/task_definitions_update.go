package cmd

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func taskDefinitionsUpdateRun(cmd *cobra.Command, args []string) {
	tdDescription, err := ecsI.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{TaskDefinition: aws.String(args[0])})
	t.Must(err)

	td := tdDescription.TaskDefinition

	var cdToUpdate *ecs.ContainerDefinition

	// If no Container Definition's name is informed, it get the first one
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
		t.Exitf("No container on the Task Family %s\n", aws.StringValue(td.Family))
	}

	if tag != "" {
		image = strings.Split(aws.StringValue(cdToUpdate.Image), ":")[0] + ":" + tag
	}

	cdToUpdate.Image = aws.String(image)

	t.Must(ecsI.RegisterTaskDefinition(&ecs.RegisterTaskDefinitionInput{
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
	}))
}

var taskDefinitionsUpdateCmd = &cobra.Command{
	Use:              "update [task-definition]",
	Short:            "Update a Task Definition",
	Args:             cobra.ExactArgs(1),
	Run:              taskDefinitionsUpdateRun,
	PersistentPreRun: persistentPreRun,
}

func init() {
	taskDefinitionsCmd.AddCommand(taskDefinitionsUpdateCmd)

	flags := taskDefinitionsUpdateCmd.Flags()

	flags.StringVar(&containerName, "container", "", containerNameSpec)
	viper.BindPFlag("container", taskDefinitionsUpdateCmd.Flags().Lookup("container"))

	flags.StringVarP(&tag, "tag", "t", "", tagSpec)
	viper.BindPFlag("tag", taskDefinitionsUpdateCmd.Flags().Lookup("tag"))

	flags.StringVarP(&image, "image", "i", "", imageSpec)
	viper.BindPFlag("image", taskDefinitionsUpdateCmd.Flags().Lookup("image"))
}

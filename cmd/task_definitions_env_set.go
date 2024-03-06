package cmd

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func taskDefinitionsEnvSetRun(cmd *cobra.Command, args []string) {
	tdDescription, err := ecsI.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(args[0]),
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
		t.Exitf("No container on the Task Family %s\n", aws.StringValue(td.Family))
	}

	env := cdToUpdate.Environment

	if taskDefinitionsEnvOverride {
		env = make([]*ecs.KeyValuePair, 0)
	}

	for _, envPair := range environmentVariables {
		envSlice := strings.Split(envPair, "=")
		if len(envSlice) != 2 {
			continue
		}

		env = append(env, &ecs.KeyValuePair{Name: aws.String(envSlice[0]), Value: aws.String(envSlice[1])})
	}

	cdToUpdate.Environment = env

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

var taskDefinitionsEnvSetCmd = &cobra.Command{
	Use:              "set [task-definition]",
	Short:            "Set a Task Definition's environment variables",
	Args:             cobra.ExactArgs(1),
	Run:              taskDefinitionsEnvSetRun,
	PersistentPreRun: persistentPreRun,
}

func init() {
	taskDefinitionsEnvCmd.AddCommand(taskDefinitionsEnvSetCmd)

	flags := taskDefinitionsEnvSetCmd.Flags()

	flags.StringVar(&containerName, "container", "", containerNameSpec)
	viper.BindPFlag("container", taskDefinitionsEnvSetCmd.Flags().Lookup("container"))

	flags.StringSliceVarP(&environmentVariables, "env", "e", []string{}, environmentVariablesSpec)
	taskDefinitionsEnvSetCmd.MarkFlagRequired("env")

	flags.BoolVar(&taskDefinitionsEnvOverride, "override", false, taskDefinitionsEnvOverrideSpec)
}

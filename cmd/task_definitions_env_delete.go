package cmd

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func taskDefinitionsEnvDeleteRun(cmd *cobra.Command, args []string) {
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

	for _, key := range environmentVariableKeys {
		for i, envKeyValuePair := range env {
			if aws.StringValue(envKeyValuePair.Name) == key {
				env = append(env[:i], env[i+1:]...)
				break
			}
		}
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

var taskDefinitionsEnvDeleteCmd = &cobra.Command{
	Use:              "delete [task-definition]",
	Short:            "Delete a Task Definition's environment variables",
	Args:             cobra.ExactArgs(1),
	Run:              taskDefinitionsEnvDeleteRun,
	PersistentPreRun: persistentPreRun,
}

func init() {
	taskDefinitionsEnvCmd.AddCommand(taskDefinitionsEnvDeleteCmd)

	flags := taskDefinitionsEnvDeleteCmd.Flags()

	flags.StringVar(&containerName, "container", "", containerNameSpec)
	viper.BindPFlag("container", taskDefinitionsEnvDeleteCmd.Flags().Lookup("container"))

	flags.StringSliceVarP(&environmentVariableKeys, "keys", "k", []string{}, environmentVariableKeysSpec)
	taskDefinitionsEnvDeleteCmd.MarkFlagRequired("keys")

	flags.BoolVar(&taskDefinitionsEnvOverride, "override", false, taskDefinitionsEnvOverrideSpec)
}

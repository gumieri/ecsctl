package cmd

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func taskDefinitionsEnvListRun(cmd *cobra.Command, args []string) {
	tdDescription, err := ecsI.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(args[0]),
	})
	t.Must(err)

	td := tdDescription.TaskDefinition

	var targetCd *ecs.ContainerDefinition

	if containerName == "" {
		targetCd = td.ContainerDefinitions[0]
	} else {
		for _, cd := range td.ContainerDefinitions {
			if aws.StringValue(cd.Name) == containerName {
				targetCd = cd
				break
			}
		}
	}

	if targetCd == nil {
		t.Exitf("No container on the Task Family %s\n", aws.StringValue(td.Family))
	}

	for _, envVar := range targetCd.Environment {
		t.Outf("%s=%s\n", aws.StringValue(envVar.Name), aws.StringValue(envVar.Value))
	}
}

var taskDefinitionsEnvListCmd = &cobra.Command{
	Use:              "list [task-definition]",
	Short:            "List a Task Definition's environment variables",
	Args:             cobra.ExactArgs(1),
	Run:              taskDefinitionsEnvListRun,
	PersistentPreRun: persistentPreRun,
}

func init() {
	taskDefinitionsEnvCmd.AddCommand(taskDefinitionsEnvListCmd)

	flags := taskDefinitionsEnvListCmd.Flags()

	flags.StringVar(&containerName, "container", "", containerNameSpec)
	viper.BindPFlag("container", taskDefinitionsEnvListCmd.Flags().Lookup("container"))
}

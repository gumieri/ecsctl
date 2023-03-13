package cmd

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/private/protocol/json/jsonutil"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func getAWSErrorCode(err error) string {
	if aerr, ok := err.(awserr.Error); ok {
		return aerr.Code()
	}
	return ""
}

func scheduledTasksConfigureRun(cmd *cobra.Command, scheduledTasks []string) {
	clustersDescription, err := ecsI.DescribeClusters(&ecs.DescribeClustersInput{
		Clusters: []*string{aws.String(viper.GetString("cluster"))},
	})
	t.Must(err)
	clusterArn := clustersDescription.Clusters[0].ClusterArn

	if len(clustersDescription.Clusters) == 0 {
		t.Exitln("Source Cluster informed not found")
	}

	ruleName := aws.String(scheduledTasks[0])
	rule, err := evbI.DescribeRule(&eventbridge.DescribeRuleInput{Name: ruleName})
	if getAWSErrorCode(err) == eventbridge.ErrCodeResourceNotFoundException {
		err = nil
		rule.EventBusName = aws.String("default")
		rule.Name = ruleName
	}
	t.Must(err)

	if c := viper.GetString("expression"); c != "" {
		rule.ScheduleExpression = aws.String(c)
	}

	switch viper.GetString("state") {
	case "disabled":
		rule.State = aws.String(eventbridge.RuleStateDisabled)
	case "enabled":
		rule.State = aws.String(eventbridge.RuleStateEnabled)
	}

	_, err = evbI.PutRule(&eventbridge.PutRuleInput{
		Description:        rule.Description,
		EventBusName:       rule.EventBusName,
		EventPattern:       rule.EventPattern,
		Name:               rule.Name,
		RoleArn:            rule.RoleArn,
		ScheduleExpression: rule.ScheduleExpression,
		State:              rule.State,
	})
	t.Must(err)

	targetsList, err := evbI.ListTargetsByRule(&eventbridge.ListTargetsByRuleInput{Rule: rule.Name})
	t.Must(err)

	var target *eventbridge.Target
	for _, t := range targetsList.Targets {
		if aws.StringValue(t.Arn) == aws.StringValue(clusterArn) {
			target = t
			break
		}
	}

	var override ecs.TaskOverride
	if target == nil {
		target = &eventbridge.Target{}
		target.Arn = clusterArn
		target.Id = rule.Name

		target.EcsParameters = &eventbridge.EcsParameters{}
		target.EcsParameters.LaunchType = aws.String("EC2")
		target.EcsParameters.TaskCount = aws.Int64(1)

		ecsEventsRole, err := iamI.GetRole(&iam.GetRoleInput{
			RoleName: aws.String(viper.GetString("events-role")),
		})
		t.Must(err)

		target.RoleArn = ecsEventsRole.Role.Arn
	} else {
		jsonutil.UnmarshalJSON(&override, strings.NewReader(aws.StringValue(target.Input)))
	}

	family := viper.GetString("task-definition")
	tdDescription, err := ecsI.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(family),
	})
	t.Must(err)

	if c := viper.GetString("command"); c != "" {
		if len(override.ContainerOverrides) == 0 {
			override.ContainerOverrides = append(override.ContainerOverrides, &ecs.ContainerOverride{
				Name: tdDescription.TaskDefinition.ContainerDefinitions[0].Name,
			})
		}
		override.ContainerOverrides[0].Command = aws.StringSlice(strings.Split(c, " "))
		bytes, err := jsonutil.BuildJSON(override)
		t.Must(err)
		target.Input = aws.String(string(bytes))
	}

	target.EcsParameters.TaskDefinitionArn = tdDescription.TaskDefinition.TaskDefinitionArn

	evbI.PutTargets(&eventbridge.PutTargetsInput{
		EventBusName: rule.EventBusName,
		Rule:         rule.Name,
		Targets:      []*eventbridge.Target{target},
	})
}

var scheduledTasksConfigureCmd = &cobra.Command{
	Use:              "configure",
	Short:            "Configure (update or create) scheduled tasks of specified cluster",
	Aliases:          []string{"c"},
	Args:             cobra.MinimumNArgs(1),
	Run:              scheduledTasksConfigureRun,
	PersistentPreRun: persistentPreRun,
}

func init() {
	scheduledTasksCmd.AddCommand(scheduledTasksConfigureCmd)

	flags := scheduledTasksConfigureCmd.Flags()

	flags.StringP("cluster", "c", "", requiredSpec+clusterSpec)
	viper.BindPFlag("cluster", scheduledTasksConfigureCmd.Flags().Lookup("cluster"))

	flags.StringP("task-definition", "T", "", requiredSpec+taskDefinitionSpec)
	viper.BindPFlag("task-definition", scheduledTasksConfigureCmd.Flags().Lookup("task-definition"))

	flags.String("command", "", "Command to override in the task execution")
	viper.BindPFlag("command", scheduledTasksConfigureCmd.Flags().Lookup("command"))

	flags.String("expression", "", "Schedule expression (Ex.: 'cron(0 3 * * ? *)' or 'rate(5 minutes)'")
	viper.BindPFlag("expression", scheduledTasksConfigureCmd.Flags().Lookup("expression"))

	flags.String("state", "", "'disabled' or 'enabled'")
	viper.BindPFlag("state", scheduledTasksConfigureCmd.Flags().Lookup("state"))

	flags.String("events-role", "ecsEventsRole", "IAM Event Role grants the AWS Event Bridge permission launch AWS Tasks")
	viper.BindPFlag("events-role", scheduledTasksConfigureCmd.Flags().Lookup("events-role"))
}

package cmd

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func scheduledTasksListRun(cmd *cobra.Command, scheduledTasks []string) {
	clustersDescription, err := ecsI.DescribeClusters(&ecs.DescribeClustersInput{
		Clusters: []*string{
			aws.String(viper.GetString("cluster")),
		},
	})

	t.Must(err)

	if len(clustersDescription.Clusters) == 0 {
		t.Exitln("Source Cluster informed not found")
	}

	list, err := evbI.ListRuleNamesByTarget(&eventbridge.ListRuleNamesByTargetInput{
		TargetArn: clustersDescription.Clusters[0].ClusterArn,
	})

	for _, name := range list.RuleNames {
		rule, err := evbI.DescribeRule(&eventbridge.DescribeRuleInput{Name: name})
		t.Must(err)

		switch state {
		case "enabled":
			if aws.StringValue(rule.State) != eventbridge.RuleStateEnabled {
				continue
			}
		case "disabled":
			if aws.StringValue(rule.State) != eventbridge.RuleStateDisabled {
				continue
			}
		}

		targets, err := evbI.ListTargetsByRule(&eventbridge.ListTargetsByRuleInput{Rule: name})
		t.Must(err)

		t.Outln(targets)
	}
}

var scheduledTasksListCmd = &cobra.Command{
	Use:              "list",
	Short:            "List scheduled tasks of specified cluster",
	Aliases:          []string{"l"},
	Run:              scheduledTasksListRun,
	PersistentPreRun: persistentPreRun,
}

func init() {
	scheduledTasksCmd.AddCommand(scheduledTasksListCmd)

	flags := scheduledTasksListCmd.Flags()

	flags.StringVarP(&state, "state", "s", "enabled", stateSpec)
	flags.StringP("cluster", "c", "", requiredSpec+clusterSpec)
	viper.BindPFlag("cluster", scheduledTasksListCmd.Flags().Lookup("cluster"))
}

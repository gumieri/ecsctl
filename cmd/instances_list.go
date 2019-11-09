package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// CompleteInstance has both ECS and EC2 instance description
type CompleteInstance struct {
	EC2 *ec2.Instance
	ECS *ecs.ContainerInstance
}

// SpotFleetRequestID get the Spot Fleet Request ID from the EC2 tags
func (i *CompleteInstance) SpotFleetRequestID() *string {
	for _, tag := range i.EC2.Tags {
		if aws.StringValue(tag.Key) == "aws:ec2spot:fleet-request-id" {
			return tag.Value
		}
	}
	return nil
}

func (i *CompleteInstance) formatProperty(property string, header bool) string {
	switch property {
	case "instance-id":
		if header {
			return "Instance ID"
		}

		return aws.StringValue(i.EC2.InstanceId)

	case "ami-id":
		if header {
			return "AMI ID"
		}

		return aws.StringValue(i.EC2.ImageId)

	case "sfr-id":
		if header {
			return "Spot Fleet request ID"
		}

		return aws.StringValue(i.SpotFleetRequestID())

	case "running-tasks":
		if header {
			return "Tasks"
		}

		return strconv.FormatInt(aws.Int64Value(i.ECS.RunningTasksCount), 10)

	case "status":
		if header {
			return "Status"
		}

		return aws.StringValue(i.ECS.Status)

	case "status-reason":
		if header {
			return "Status reason"
		}

		return aws.StringValue(i.ECS.StatusReason)

	case "agent-connected":
		if header {
			return "agent connected?"
		}

		if aws.BoolValue(i.ECS.AgentConnected) {
			return "Yes"
		}

		return "No"

	case "launch-time":
		if header {
			return "Status"
		}

		return aws.TimeValue(i.EC2.LaunchTime).Format("2006-01-02T15:04:05")

	}

	return ""
}

func instancesListRun(cmd *cobra.Command, instances []string) {
	clustersDescription, err := ecsI.DescribeClusters(&ecs.DescribeClustersInput{
		Clusters: []*string{
			aws.String(viper.GetString("cluster")),
		},
	})

	t.Must(err)

	if len(clustersDescription.Clusters) == 0 {
		t.Exitln("Source Cluster informed not found")
	}

	c := clustersDescription.Clusters[0]

	instancesList, err := ecsI.ListContainerInstances(&ecs.ListContainerInstancesInput{
		Cluster: c.ClusterName,
	})

	t.Must(err)

	instancesDescription, err := ecsI.DescribeContainerInstances(&ecs.DescribeContainerInstancesInput{
		Cluster:            c.ClusterName,
		ContainerInstances: instancesList.ContainerInstanceArns,
	})
	t.Must(err)

	ec2InstancesIDs := make([]*string, 0)
	completeInstances := make([]*CompleteInstance, 0)
	for _, instance := range instancesDescription.ContainerInstances {
		ec2InstancesIDs = append(ec2InstancesIDs, instance.Ec2InstanceId)
		completeInstances = append(completeInstances, &CompleteInstance{ECS: instance})
	}

	ec2Description, err := ec2I.DescribeInstances(&ec2.DescribeInstancesInput{InstanceIds: ec2InstancesIDs})
	t.Must(err)

	for _, reservation := range ec2Description.Reservations {
		for _, instance := range reservation.Instances {
			for _, completeInstance := range completeInstances {
				if aws.StringValue(completeInstance.ECS.Ec2InstanceId) != aws.StringValue(instance.InstanceId) {
					continue
				}

				completeInstance.EC2 = instance
			}
		}
	}

	properties := viper.GetStringSlice("properties")

	header := []string{}
	for _, p := range properties {
		header = append(header, completeInstances[0].formatProperty(p, true))
	}

	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, strings.Join(header, "\t"))

	for _, i := range completeInstances {
		line := []string{}
		for _, p := range properties {
			line = append(line, i.formatProperty(p, false))
		}
		fmt.Fprintln(w, strings.Join(line, "\t"))
	}

	w.Flush()
}

var instancesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List instances of specified cluster",
	Run:   instancesListRun,
}

func init() {
	instancesCmd.AddCommand(instancesListCmd)

	flags := instancesListCmd.Flags()

	flags.StringP("cluster", "c", "", requiredSpec+clusterSpec)
	viper.BindPFlag("cluster", instancesListCmd.Flags().Lookup("cluster"))

	defaultColumns := []string{
		"instance-id",
		"running-tasks",
		"status",
		"agent-connected",
	}

	flags.StringSliceP("properties", "p", defaultColumns, "properties to be listed as column\n")

	viper.BindPFlag("properties", instancesListCmd.Flags().Lookup("properties"))
}

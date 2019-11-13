package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/ahmetb/go-cursor"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/gumieri/typist"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// CompleteInstance has both ECS and EC2 instance description
type CompleteInstance struct {
	EC2 *ec2.Instance
	ECS *ecs.ContainerInstance
}

// GetInstances return a slice of CompleteInstance from specified Cluster
func GetInstances(c *ecs.Cluster) ([]*CompleteInstance, error) {
	instancesList, err := ecsI.ListContainerInstances(&ecs.ListContainerInstancesInput{
		Cluster: c.ClusterName,
	})

	if err != nil {
		return nil, err
	}

	instancesDescription, err := ecsI.DescribeContainerInstances(&ecs.DescribeContainerInstancesInput{
		Cluster:            c.ClusterName,
		ContainerInstances: instancesList.ContainerInstanceArns,
	})

	if err != nil {
		return nil, err
	}

	ec2InstancesIDs := make([]*string, 0)
	completeInstances := make([]*CompleteInstance, 0)
	for _, instance := range instancesDescription.ContainerInstances {
		ec2InstancesIDs = append(ec2InstancesIDs, instance.Ec2InstanceId)
		completeInstances = append(completeInstances, &CompleteInstance{ECS: instance})
	}

	ec2Description, err := ec2I.DescribeInstances(&ec2.DescribeInstancesInput{InstanceIds: ec2InstancesIDs})

	if err != nil {
		return nil, err
	}

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

	return completeInstances, nil
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

	case "agent-version":
		if header {
			return "Agent version"
		}

		return aws.StringValue(i.ECS.VersionInfo.AgentVersion)

	case "docker-version":
		if header {
			return "Docker version"
		}

		return aws.StringValue(i.ECS.VersionInfo.DockerVersion)
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

	watch := viper.GetBool("watch")
	interval := viper.GetInt("interval")
	if interval != 0 {
		watch = true
	} else if watch {
		interval = 5
	}

	lc := 0
	for {
		completeInstances, err := GetInstances(clustersDescription.Clusters[0])
		t.Must(err)

		properties := viper.GetStringSlice("properties")

		table := &typist.Table{
			Header:   []string{},
			Lines:    [][]string{},
			MinWidth: 1,
			TabWidth: 1,
			Padding:  1,
			PadChar:  ' ',
		}

		for _, p := range properties {
			table.Header = append(table.Header, completeInstances[0].formatProperty(p, true))
		}

		for _, instance := range completeInstances {
			line := []string{}
			for _, p := range properties {
				line = append(line, instance.formatProperty(p, false))
			}
			table.Lines = append(table.Lines, line)
		}

		t.Config.Header = !viper.GetBool("no-header")

		if lc != 0 {
			fmt.Print(cursor.MoveUp(lc))
		}
		t.Table(table)

		lc = len(table.Lines)
		if t.Config.Header {
			lc++
		}

		fmt.Print(cursor.ClearScreenDown())

		if !watch {
			t.Exit(nil)
		}

		time.Sleep(time.Duration(interval) * time.Second)
	}
}

var instancesListCmd = &cobra.Command{
	Use:              "list",
	Short:            "List instances of specified cluster",
	Aliases:          []string{"l"},
	Run:              instancesListRun,
	PersistentPreRun: persistentPreRun,
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

	flags.Bool("no-header", false, noHeaderSpec)
	viper.BindPFlag("no-header", instancesListCmd.Flags().Lookup("no-header"))

	flags.Bool("watch", false, watchSpec)
	viper.BindPFlag("watch", instancesListCmd.Flags().Lookup("watch"))

	flags.Int("interval", 0, "Interval of time (in seconds) refreshing (default: 5)")
	viper.BindPFlag("interval", instancesListCmd.Flags().Lookup("interval"))

	viper.BindPFlag("properties", instancesListCmd.Flags().Lookup("properties"))
}

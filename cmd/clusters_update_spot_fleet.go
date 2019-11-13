package cmd

import (
	"errors"
	"fmt"
	"time"

	"github.com/ahmetb/go-cursor"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/gumieri/typist"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func clustersUpdateSpotFleetRun(cmd *cobra.Command, clusters []string) {
	clustersDescription, err := ecsI.DescribeClusters(&ecs.DescribeClustersInput{
		Clusters: []*string{
			aws.String(clusters[0]),
		},
	})
	t.Must(err)

	if len(clustersDescription.Clusters) == 0 {
		t.Must(errors.New("Cluster informed not found"))
	}

	c := clustersDescription.Clusters[0]
	instances, err := GetInstances(c)
	t.Must(err)

	oldSFRID := instances[0].SpotFleetRequestID()
	sfrDescription, err := ec2I.DescribeSpotFleetRequests(&ec2.DescribeSpotFleetRequestsInput{
		SpotFleetRequestIds: []*string{oldSFRID},
	})
	t.Must(err)

	oldSFR := sfrDescription.SpotFleetRequestConfigs[0]

	LaunchSpecifications := oldSFR.SpotFleetRequestConfig.LaunchSpecifications

	ami := viper.GetString("ami")
	if ami != "" {
		if ami == "latest" {
			platform := aws.StringValue(instances[0].EC2.Platform)
			latestAMI, err := latestAmiEcsOptimized(platform)
			t.Must(err)
			ami = aws.StringValue(latestAMI.ImageId)
		}

		for _, ls := range LaunchSpecifications {
			ls.ImageId = aws.String(ami)
		}
	}

	SpotFleetRequestConfig := ec2.SpotFleetRequestConfigData{
		IamFleetRole:         oldSFR.SpotFleetRequestConfig.IamFleetRole,
		LaunchSpecifications: LaunchSpecifications,
		TargetCapacity:       oldSFR.SpotFleetRequestConfig.TargetCapacity,
		AllocationStrategy:   oldSFR.SpotFleetRequestConfig.AllocationStrategy,
	}

	rsfo, err := ec2I.RequestSpotFleet(&ec2.RequestSpotFleetInput{
		SpotFleetRequestConfig: &SpotFleetRequestConfig,
	})
	t.Must(err)

	newSFRID := rsfo.SpotFleetRequestId

	lc := 0
	cancelRequest := false
	drainRequested := false
	for {
		completeInstances, err := GetInstances(c)
		t.Must(err)

		properties := []string{
			"instance-id",
			"running-tasks",
			"status",
			"agent-connected",
			"agent-version",
			"sfr-id",
		}

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

		var oldTasksCount int64
		var oldInstancesCount int
		var oldContainerInstances []*string
		for _, instance := range completeInstances {
			line := []string{}
			for _, p := range properties {
				line = append(line, instance.formatProperty(p, false))
			}
			table.Lines = append(table.Lines, line)

			if aws.StringValue(instance.SpotFleetRequestID()) == aws.StringValue(oldSFRID) {
				oldTasksCount += aws.Int64Value(instance.ECS.RunningTasksCount)
				oldInstancesCount++
				oldContainerInstances = append(oldContainerInstances, instance.ECS.ContainerInstanceArn)
			}
		}

		if lc != 0 {
			fmt.Print(cursor.MoveUp(lc))
		}
		t.Table(table)

		lc = len(table.Lines) + 1

		fmt.Print(cursor.ClearScreenDown())

		if !drainRequested {
			dsfro, err := ec2I.DescribeSpotFleetRequests(&ec2.DescribeSpotFleetRequestsInput{
				SpotFleetRequestIds: []*string{newSFRID},
			})
			t.Must(err)

			sfrc := dsfro.SpotFleetRequestConfigs[0]
			if aws.StringValue(sfrc.ActivityStatus) == "fulfilled" {
				t.Must(ecsI.UpdateContainerInstancesState(&ecs.UpdateContainerInstancesStateInput{
					Cluster:            aws.String(clusters[0]),
					ContainerInstances: oldContainerInstances,
					Status:             aws.String("DRAINING"),
				}))
				drainRequested = true
			}
		}

		if oldTasksCount <= 0 && !cancelRequest {
			t.Must(ec2I.CancelSpotFleetRequests(&ec2.CancelSpotFleetRequestsInput{
				SpotFleetRequestIds: []*string{oldSFRID},
				TerminateInstances:  aws.Bool(true),
			}))
			cancelRequest = true
		}

		if oldInstancesCount <= 0 {
			t.Exit(nil)
		}

		time.Sleep(5 * time.Second)
	}
}

var clustersUpdateSpotFleetCmd = &cobra.Command{
	Use:              "update-spot-fleet [cluster]",
	Short:            "Create a new Spot Fleet request from a previous one and wait for the previous draining before cancelling.",
	Aliases:          []string{"update-spotfleet", "update-sf", "updatesf", "usf"},
	Run:              clustersUpdateSpotFleetRun,
	PersistentPreRun: persistentPreRun,
}

func init() {
	clustersCmd.AddCommand(clustersUpdateSpotFleetCmd)

	flags := clustersUpdateSpotFleetCmd.Flags()
	flags.String("ami", "", `The Amazon EC2 Image (ID or tag 'Name')
The "latest" refers to the latest ECS Optimized published by the AWS`)
	viper.BindPFlag("ami", clustersUpdateSpotFleetCmd.Flags().Lookup("ami"))

	// flags.StringSliceP("subnet", "n", []string{}, subnetsSpec)
	// viper.BindPFlag("subnet", clustersUpdateSpotFleetCmd.Flags().Lookup("subnet"))

	// flags.StringSliceP("instance-type", "i", []string{}, instanceTypesSpec)
	// viper.BindPFlag("instance-type", clustersUpdateSpotFleetCmd.Flags().Lookup("instance-type"))

	// flags.StringSliceP("security-group", "g", []string{}, securityGroupsSpec)
	// viper.BindPFlag("security-group", clustersUpdateSpotFleetCmd.Flags().Lookup("security-group"))

	// flags.Int64P("target-capacity", "c", 1, targetCapacitySpec)
	// viper.BindPFlag("target-capacity", clustersUpdateSpotFleetCmd.Flags().Lookup("target-capacity"))

	// flags.StringP("allocation-strategy", "s", "", allocationStrategySpec)
	// viper.BindPFlag("allocation-strategy", clustersUpdateSpotFleetCmd.Flags().Lookup("allocation-strategy"))

	// flags.String("spot-price", "", spotPriceSpec)
	// viper.BindPFlag("spot-price", clustersUpdateSpotFleetCmd.Flags().Lookup("spot-price"))

	// flags.String("kernel-id", "", kernelIDSpec)
	// viper.BindPFlag("kernel-id", clustersUpdateSpotFleetCmd.Flags().Lookup("kernel-id"))

	// flags.StringP("key", "k", "", keySpec)
	// viper.BindPFlag("key", clustersUpdateSpotFleetCmd.Flags().Lookup("key"))

	// flags.StringSliceP("tag", "t", []string{}, tagsSpec)
	// viper.BindPFlag("tag", clustersUpdateSpotFleetCmd.Flags().Lookup("tag"))
}

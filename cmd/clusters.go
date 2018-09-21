package cmd

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/cobra"
)

type templateUserData struct {
	Cluster string
	Region  string
}

func parseTags(s string) (parsed []*ec2.Tag) {
	for _, kv := range strings.Split(s, ",") {
		kvs := strings.Split(kv, "=")
		if len(kvs) != 2 {
			continue
		}

		tag := ec2.Tag{
			Key:   aws.String(kvs[0]),
			Value: aws.String(kvs[1]),
		}

		parsed = append(parsed, &tag)
	}
	return
}

func latestAmiEcsOptimized() (latestImage ec2.Image, err error) {
	result, err := ec2I.DescribeImages(&ec2.DescribeImagesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("state"),
				Values: []*string{aws.String("available")},
			},
			{
				// Amazon ECS Images
				Name:   aws.String("owner-alias"),
				Values: []*string{aws.String("amazon")},
			},
			{
				Name:   aws.String("name"),
				Values: []*string{aws.String("amzn-ami-?????????-amazon-ecs-optimized")},
			},
		},
	})

	if err != nil {
		return
	}

	for _, image := range result.Images {
		if aws.StringValue(latestImage.CreationDate) < aws.StringValue(image.CreationDate) {
			latestImage = *image
			continue
		}
	}

	return
}

func findSecurityGroup(s string) (securityGroup *ec2.SecurityGroup, err error) {
	sgd, err := ec2I.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("group-id"),
				Values: []*string{aws.String(s)},
			},
		},
	})

	if err != nil {
		return
	}

	if len(sgd.SecurityGroups) > 0 {
		securityGroup = sgd.SecurityGroups[0]
		return
	}

	sgd, err = ec2I.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("group-name"),
				Values: []*string{aws.String(s)},
			},
		},
	})

	if len(sgd.SecurityGroups) > 0 {
		securityGroup = sgd.SecurityGroups[0]
		return
	}

	sgd, err = ec2I.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("tag:Name"),
				Values: []*string{aws.String(s)},
			},
		},
	})

	if len(sgd.SecurityGroups) == 0 {
		err = fmt.Errorf("SecurityGroup (%s) not found", s)
		return
	}

	securityGroup = sgd.SecurityGroups[0]

	return
}

func findSubnet(s string) (subnet *ec2.Subnet, err error) {
	sd, err := ec2I.DescribeSubnets(&ec2.DescribeSubnetsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("subnet-id"),
				Values: []*string{aws.String(s)},
			},
		},
	})

	if err != nil {
		return
	}

	if len(sd.Subnets) > 0 {
		subnet = sd.Subnets[0]
		return
	}

	sd, err = ec2I.DescribeSubnets(&ec2.DescribeSubnetsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("tag:Name"),
				Values: []*string{aws.String(s)},
			},
		},
	})

	if err != nil {
		return
	}

	if len(sd.Subnets) == 0 {
		err = fmt.Errorf("Subnet (%s) not found", s)
		return
	}

	subnet = sd.Subnets[0]

	return
}

func clustersRun(cmd *cobra.Command, args []string) {
	cmd.Help()
}

var clustersCmd = &cobra.Command{
	Use:   "clusters [command]",
	Short: "commands to manage clusters",
	Run:   clustersRun,
}

func init() {
	rootCmd.AddCommand(clustersCmd)
}

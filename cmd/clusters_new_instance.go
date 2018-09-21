package cmd

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/spf13/cobra"
)

var ec2InstanceUserData = `
#!/bin/bash
echo ECS_CLUSTER={{.Cluster}} >> /etc/ecs/ecs.config;echo ECS_BACKEND_HOST= >> /etc/ecs/ecs.config;
`

func clustersNewInstanceRun(cmd *cobra.Command, clusters []string) {
	clustersDescription, err := ecsI.DescribeClusters(&ecs.DescribeClustersInput{
		Clusters: []*string{
			aws.String(clusters[0]),
		},
	})
	must(err)

	if len(clustersDescription.Clusters) == 0 {
		must(errors.New("Cluster informed not found"))
	}

	c := clustersDescription.Clusters[0]

	tmpl, err := template.New("UserData").Parse(ec2InstanceUserData)
	must(err)

	userDataF := new(bytes.Buffer)
	must(tmpl.Execute(userDataF, templateUserData{Cluster: *c.ClusterName}))

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	latestImage, err := latestAmiEcsOptimized()
	must(err)

	// TODO: automaticaly --create-roles if does not exist
	if instanceRole == "" {
		instanceRole = "ecsInstanceRole"
	}

	instanceRoleResponse, err := iamI.GetRole(&iam.GetRoleInput{
		RoleName: aws.String(instanceRole),
	})
	must(err)

	subnetDescription, err := findSubnet(subnet)
	must(err)

	// TODO: AWS Tags
	RunInstancesInput := ec2.RunInstancesInput{
		IamInstanceProfile: &ec2.IamInstanceProfileSpecification{Name: instanceRoleResponse.Role.RoleName},
		EbsOptimized:       aws.Bool(ebs),
		ImageId:            latestImage.ImageId,
		SubnetId:           subnetDescription.SubnetId,
		InstanceType:       aws.String(instanceType),
		UserData:           aws.String(base64.StdEncoding.EncodeToString(userDataF.Bytes())),
		MinCount:           aws.Int64(minimum),
		MaxCount:           aws.Int64(maximum),
	}

	if tags != "" {
		RunInstancesInput.TagSpecifications = []*ec2.TagSpecification{
			&ec2.TagSpecification{
				ResourceType: aws.String("instance"),
				Tags:         parseTags(tags),
			},
		}
	}

	if credit != "" {
		RunInstancesInput.CreditSpecification = &ec2.CreditSpecificationRequest{
			CpuCredits: aws.String(credit),
		}
	}

	if securityGroups != "" {
		var sgs []*string
		for _, securityGroup := range strings.Split(securityGroups, ",") {
			sg, err := findSecurityGroup(securityGroup)
			must(err)
			sgs = append(sgs, sg.GroupId)
		}
		RunInstancesInput.SecurityGroupIds = sgs
	}

	if kernelID != "" {
		RunInstancesInput.KernelId = aws.String(kernelID)
	}

	if key != "" {
		RunInstancesInput.KeyName = aws.String(key)
	}

	if monitoring {
		RunInstancesInput.Monitoring = &ec2.RunInstancesMonitoringEnabled{Enabled: aws.Bool(monitoring)}
	}

	_, err = ec2I.RunInstances(&RunInstancesInput)
	must(err)
}

var clustersNewInstanceCmd = &cobra.Command{
	Use:   "new-instance [cluster]",
	Short: "Add a new EC2 instance to informed cluster",
	Args:  cobra.ExactArgs(1),
	Run:   clustersNewInstanceRun,
}

func init() {
	clustersCmd.AddCommand(clustersNewInstanceCmd)

	flags := clustersNewInstanceCmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&instanceType, "instance-type", "", requiredSpec+instanceTypeSpec)
	flags.StringVar(&subnet, "subnet", "", requiredSpec+subnetSpec)
	flags.StringVar(&securityGroups, "security-groups", "", securityGroupsSpec)

	flags.StringVar(&key, "key", "", keySpec)
	flags.StringVar(&tags, "tags", "", tagsSpec)
	flags.Int64Var(&minimum, "min", 1, minimumSpec)
	flags.Int64Var(&maximum, "max", 1, maximumSpec)
	flags.StringVar(&credit, "credit", "", creditSpec)

	clustersNewSpotFleetCmd.MarkFlagRequired("subnet")
	clustersNewSpotFleetCmd.MarkFlagRequired("instance-type")
}

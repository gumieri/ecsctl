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

var instanceType string
var subnet string
var credit string
var minimum int64
var maximum int64

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

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if len(clustersDescription.Clusters) == 0 {
		fmt.Println(errors.New("Cluster informed not found"))
		os.Exit(1)
	}

	c := clustersDescription.Clusters[0]

	tmpl, err := template.New("UserData").Parse(ec2InstanceUserData)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	userDataF := new(bytes.Buffer)
	err = tmpl.Execute(userDataF, templateUserData{
		Cluster: *c.ClusterName,
	})

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	latestImage, err := latestAmiEcsOptimized()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// TODO: automaticaly --create-roles if does not exist
	if instanceRole == "" {
		instanceRole = "ecsInstanceRole"
	}

	instanceRoleResponse, err := iamI.GetRole(&iam.GetRoleInput{
		RoleName: aws.String(instanceRole),
	})

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// TODO: AWS Tags
	RunInstancesInput := ec2.RunInstancesInput{
		IamInstanceProfile: &ec2.IamInstanceProfileSpecification{Name: instanceRoleResponse.Role.RoleName},
		EbsOptimized:       aws.Bool(ebs),
		ImageId:            latestImage.ImageId,
		SubnetId:           aws.String(subnet),
		InstanceType:       aws.String(instanceType),
		UserData:           aws.String(base64.StdEncoding.EncodeToString(userDataF.Bytes())),
		MinCount:           aws.Int64(minimum),
		MaxCount:           aws.Int64(maximum),
	}

	if credit != "" {
		RunInstancesInput.CreditSpecification = &ec2.CreditSpecificationRequest{
			CpuCredits: aws.String(credit),
		}
	}

	if securityGroups != "" {
		var sgs []*string
		for _, value := range strings.Split(securityGroups, ",") {
			sg := value
			sgs = append(sgs, &sg)
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
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
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
	flags.StringVar(&instanceType, "instance-type", "", "REQUIRED - Type of instance to be launched.")
	flags.StringVar(&key, "key", "", "Key name to access the instances.")
	flags.StringVar(&subnet, "subnet", "", "REQUIRED - The Subnet ID to launch the instance.")
	flags.StringVar(&securityGroups, "security-groups", "", "Security Groups for the instances (separeted by comma ',').")

	flags.StringVar(&credit, "credit", "", "The credit option for CPU usage of a T2 or T3 instance (valid values: 'standard' or 'unlimited').")

	flags.Int64Var(&minimum, "min", 1, "The minimum number of instances to launch. If you specify a minimum that is more instances than Amazon EC2 can launch in the target Availability Zone, Amazon EC2 launches no instances (default: 1).")

	flags.Int64Var(&maximum, "max", 1, "The maximum number of instances to launch. If you specify more instances than Amazon EC2 can launch in the target Availability Zone, Amazon EC2 launches the largest possible number of instances above MinCount (default: 1).")

	clustersNewSpotFleetCmd.MarkFlagRequired("instance-type")
	clustersNewSpotFleetCmd.MarkFlagRequired("subnet")
}

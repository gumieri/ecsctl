package cmd

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"os"

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

func clustersAddInstanceRun(cmd *cobra.Command, clusters []string) {
	clustersDescription, err := ecsI.DescribeClusters(&ecs.DescribeClustersInput{
		Clusters: []*string{
			aws.String(clusters[0]),
		},
	})
	typist.Must(err)

	if len(clustersDescription.Clusters) == 0 {
		typist.Must(errors.New("Cluster informed not found"))
	}

	c := clustersDescription.Clusters[0]

	tmpl, err := template.New("UserData").Parse(ec2InstanceUserData)
	typist.Must(err)

	userDataF := new(bytes.Buffer)
	typist.Must(tmpl.Execute(userDataF, templateUserData{Cluster: *c.ClusterName}))

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	latestImage, err := latestAmiEcsOptimized()
	typist.Must(err)

	// TODO: automaticaly --create-roles if does not exist
	if instanceRole == "" {
		instanceRole = "ecsInstanceRole"
	}

	instanceRoleResponse, err := iamI.GetRole(&iam.GetRoleInput{
		RoleName: aws.String(instanceRole),
	})
	typist.Must(err)

	subnetDescription, err := findSubnet(subnet)
	typist.Must(err)

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

	if len(tags) > 0 {
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

	var sgs []*string
	for _, securityGroup := range securityGroups {
		sg, err := findSecurityGroup(securityGroup)
		typist.Must(err)
		sgs = append(sgs, sg.GroupId)
	}
	RunInstancesInput.SecurityGroupIds = sgs

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
	typist.Must(err)
}

var clustersAddInstanceCmd = &cobra.Command{
	Use:   "add-instance [cluster]",
	Short: "Add a add EC2 instance to informed cluster",
	Args:  cobra.ExactArgs(1),
	Run:   clustersAddInstanceRun,
}

func init() {
	clustersCmd.AddCommand(clustersAddInstanceCmd)

	flags := clustersAddInstanceCmd.Flags()
	flags.SortFlags = false

	flags.StringVarP(&instanceType, "instance-type", "i", "", requiredSpec+instanceTypeSpec)
	flags.StringVarP(&subnet, "subnet", "n", "", requiredSpec+subnetSpec)
	flags.StringSliceVarP(&securityGroups, "security-groups", "g", []string{}, securityGroupsSpec)

	flags.StringVarP(&key, "key", "k", "", keySpec)
	flags.StringSliceVarP(&tags, "tag", "t", []string{}, tagsSpec)
	flags.Int64Var(&minimum, "min", 1, minimumSpec)
	flags.Int64Var(&maximum, "max", 1, maximumSpec)
	flags.StringVar(&credit, "credit", "", creditSpec)

	clustersAddSpotFleetCmd.MarkFlagRequired("subnet")
	clustersAddSpotFleetCmd.MarkFlagRequired("instance-type")
}

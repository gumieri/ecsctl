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

var spotPrice string
var spotFleetRole string
var instanceRole string
var targetCapacity int64
var allocationStrategy string
var instanceTypes string
var securityGroups string
var subnets string
var key string

var userData = `
#!/bin/bash
echo ECS_CLUSTER={{.Cluster}} >> /etc/ecs/ecs.config
echo ECS_BACKEND_HOST= >> /etc/ecs/ecs.config
export PATH=/usr/local/bin:$PATH
yum -y install jq
easy_install pip
pip install awscli
aws configure set default.region us-east-1
cat <<EOF > /etc/init/spot-instance-termination-notice-handler.conf
description "Start spot instance termination handler monitoring script"
author "Amazon Web Services"
start on started ecs
script
echo \$\$ > /var/run/spot-instance-termination-notice-handler.pid
exec /usr/local/bin/spot-instance-termination-notice-handler.sh
end script
pre-start script
logger "[spot-instance-termination-notice-handler.sh]: spot instance termination
notice handler started"
end script
EOF
cat <<EOF > /usr/local/bin/spot-instance-termination-notice-handler.sh
#!/bin/bash
while sleep 5; do
  if [ -z \$(curl -Isf http://169.254.169.254/latest/meta-data/spot/termination-time)]; then
    /bin/false
  else
    logger "[spot-instance-termination-notice-handler.sh]: spot instance termination notice detected"
    STATUS=DRAINING
    ECS_CLUSTER=\$(curl -s http://localhost:51678/v1/metadata | jq .Cluster | tr -d \")
    CONTAINER_INSTANCE=\$(curl -s http://localhost:51678/v1/metadata | jq .ContainerInstanceArn | tr -d \")
    logger "[spot-instance-termination-notice-handler.sh]: putting instance in state \$STATUS"

    /usr/local/bin/aws  ecs update-container-instances-state --cluster \$ECS_CLUSTER --container-instances \$CONTAINER_INSTANCE --status \$STATUS

    logger "[spot-instance-termination-notice-handler.sh]: putting myself to sleep..."
    sleep 120 # exit loop as instance expires in 120 secs after terminating notification
  fi
done
EOF
chmod +x /usr/local/bin/spot-instance-termination-notice-handler.sh
`

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

type templateData struct {
	Cluster string
}

func clustersNewSpotFleetRun(cmd *cobra.Command, clusters []string) {
	clustersDescription, err := ecsI.DescribeClusters(&ecs.DescribeClustersInput{
		Clusters: []*string{
			aws.String(clusters[0]),
		},
	})

	if len(clustersDescription.Clusters) == 0 {
		fmt.Println(errors.New("Cluster informed not found"))
		os.Exit(1)
	}

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	c := clustersDescription.Clusters[0]

	tmpl, err := template.New("UserData").Parse(userData)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	userDataF := new(bytes.Buffer)
	err = tmpl.Execute(userDataF, templateData{
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

	if spotFleetRole == "" {
		spotFleetRole = "ecsSpotFleetRole"
	}

	spotFleetRoleResponse, err := iamI.GetRole(&iam.GetRoleInput{
		RoleName: aws.String(spotFleetRole),
	})

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

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

	var SecurityGroups []*ec2.GroupIdentifier
	for _, securityGroup := range strings.Split(securityGroups, ",") {
		SecurityGroups = append(SecurityGroups, &ec2.GroupIdentifier{
			GroupId: aws.String(securityGroup),
		})
	}

	var LaunchSpecifications []*ec2.SpotFleetLaunchSpecification
	for _, instanceType := range strings.Split(instanceTypes, ",") {
		LaunchSpecifications = append(LaunchSpecifications, &ec2.SpotFleetLaunchSpecification{
			IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
				Name: instanceRoleResponse.Role.RoleName,
			},
			ImageId:        latestImage.ImageId,
			InstanceType:   aws.String(instanceType),
			KeyName:        aws.String(key),
			SecurityGroups: SecurityGroups,
			SubnetId:       aws.String(subnets),
			UserData:       aws.String(base64.StdEncoding.EncodeToString(userDataF.Bytes())),
		})
	}

	SpotFleetRequestConfig := ec2.SpotFleetRequestConfigData{
		IamFleetRole:         spotFleetRoleResponse.Role.Arn,
		LaunchSpecifications: LaunchSpecifications,
		SpotPrice:            aws.String(spotPrice),
		TargetCapacity:       aws.Int64(targetCapacity),
	}

	if allocationStrategy != "" {
		SpotFleetRequestConfig.AllocationStrategy = aws.String(allocationStrategy)
	}

	_, err = ec2I.RequestSpotFleet(&ec2.RequestSpotFleetInput{
		SpotFleetRequestConfig: &SpotFleetRequestConfig,
	})

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

var clustersNewSpotFleetCmd = &cobra.Command{
	Use:   "new-spot-fleet [cluster]",
	Short: "Add a new Spot Fleet to informed cluster",
	Args:  cobra.ExactArgs(1),
	Run:   clustersNewSpotFleetRun,
}

func init() {
	clustersCmd.AddCommand(clustersNewSpotFleetCmd)

	clustersNewSpotFleetCmd.Flags().StringVar(&key, "key", "", "Key name to access the instances.")
	clustersNewSpotFleetCmd.MarkFlagRequired("key")

	clustersNewSpotFleetCmd.Flags().StringVar(&spotPrice, "spot-price", "", "Top price to pay for the spot instances.")
	clustersNewSpotFleetCmd.MarkFlagRequired("spot-price")

	clustersNewSpotFleetCmd.Flags().Int64Var(&targetCapacity, "target-capacity", 0, "The capacity amout defined for the cluster.")
	clustersNewSpotFleetCmd.MarkFlagRequired("target-capacity")

	clustersNewSpotFleetCmd.Flags().StringVar(&instanceTypes, "instance-types", "", "Types of instance to be used by the Spot Fleet (separeted by comma ',').")
	clustersNewSpotFleetCmd.MarkFlagRequired("instance-types")

	clustersNewSpotFleetCmd.Flags().StringVar(&securityGroups, "security-groups", "", "Security Groups for the instances (separeted by comma ',').")
	clustersNewSpotFleetCmd.MarkFlagRequired("security-groups")

	clustersNewSpotFleetCmd.Flags().StringVar(&subnets, "subnets", "", "Type of instance to be used by the Spot Fleet (separeted by comma ',').")
	clustersNewSpotFleetCmd.MarkFlagRequired("subnets")

	clustersNewSpotFleetCmd.Flags().StringVar(&allocationStrategy, "allocation-strategy", "", "diversified or lowestPrice (default: lowestPrice).")

	clustersNewSpotFleetCmd.Flags().StringVar(&spotFleetRole, "spot-fleet-role", "", "IAM fleet role grants the Spot fleet permission launch and terminate instances on your behalf.")

	clustersNewSpotFleetCmd.Flags().StringVar(&spotFleetRole, "spot-fleet-role", "", "IAM fleet role grants the Spot fleet permission launch and terminate instances on your behalf.")

	clustersNewSpotFleetCmd.Flags().StringVar(&instanceRole, "instance-role", "", "An instance profile is a container for an IAM role and enables you to pass role information to Amazon EC2 Instance when the instance starts.")
}

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

var spotFleetUserData = `
#!/bin/bash
echo ECS_CLUSTER={{.Cluster}} >> /etc/ecs/ecs.config
echo ECS_BACKEND_HOST= >> /etc/ecs/ecs.config
export PATH=/usr/local/bin:$PATH
yum -y install jq
easy_install pip
pip install awscli
aws configure set default.region {{.Region}}
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

func clustersNewSpotFleetRun(cmd *cobra.Command, clusters []string) {
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

	tmpl, err := template.New("UserData").Parse(spotFleetUserData)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	userDataF := new(bytes.Buffer)
	err = tmpl.Execute(userDataF, templateUserData{
		Cluster: *c.ClusterName,
		Region:  aws.StringValue(awsSession.Config.Region),
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

	var SecurityGroups []*ec2.GroupIdentifier
	for _, securityGroup := range strings.Split(securityGroups, ",") {
		sg, err := findSecurityGroup(securityGroup)
		must(err)

		SecurityGroups = append(SecurityGroups, &ec2.GroupIdentifier{
			GroupId: sg.GroupId,
		})
	}

	var subnetsIds []string
	for _, subnet := range strings.Split(subnets, ",") {
		Subnet, err := findSubnet(subnet)
		must(err)
		subnetsIds = append(subnetsIds, aws.StringValue(Subnet.SubnetId))
	}

	var LaunchSpecifications []*ec2.SpotFleetLaunchSpecification
	for _, instanceType := range strings.Split(instanceTypes, ",") {
		SpotFleetLaunchSpecification := ec2.SpotFleetLaunchSpecification{
			IamInstanceProfile: &ec2.IamInstanceProfileSpecification{Name: instanceRoleResponse.Role.RoleName},
			EbsOptimized:       aws.Bool(ebs),
			ImageId:            latestImage.ImageId,
			InstanceType:       aws.String(instanceType),
			SecurityGroups:     SecurityGroups,
			SubnetId:           aws.String(strings.Join(subnetsIds, ",")),
			UserData:           aws.String(base64.StdEncoding.EncodeToString(userDataF.Bytes())),
		}

		if kernelID != "" {
			SpotFleetLaunchSpecification.KernelId = aws.String(kernelID)
		}

		if key != "" {
			SpotFleetLaunchSpecification.KeyName = aws.String(key)
		}

		if monitoring {
			SpotFleetLaunchSpecification.Monitoring = &ec2.SpotFleetMonitoring{Enabled: aws.Bool(monitoring)}
		}

		if tags != "" {
			SpotFleetLaunchSpecification.TagSpecifications = []*ec2.SpotFleetTagSpecification{
				&ec2.SpotFleetTagSpecification{
					ResourceType: aws.String("instance"),
					Tags:         parseTags(tags),
				},
			}
		}

		LaunchSpecifications = append(LaunchSpecifications, &SpotFleetLaunchSpecification)
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

	flags := clustersNewSpotFleetCmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&subnets, "subnets", "", requiredSpec+subnetsSpec)
	flags.StringVar(&spotPrice, "spot-price", "", requiredSpec+spotPriceSpec)
	flags.StringVar(&instanceTypes, "instance-types", "", requiredSpec+instanceTypesSpec)
	flags.StringVar(&securityGroups, "security-groups", "", requiredSpec+securityGroupsSpec)
	flags.Int64Var(&targetCapacity, "target-capacity", 1, targetCapacitySpec)
	flags.StringVar(&instanceRole, "instance-role", "", instanceRoleSpec)
	flags.StringVar(&spotFleetRole, "spot-fleet-role", "", spotFleetRoleSpec)
	flags.StringVar(&allocationStrategy, "allocation-strategy", "", allocationStrategySpec)
	flags.BoolVar(&monitoring, "monitoring", false, monitoringSpec)
	flags.StringVar(&kernelID, "kernel-id", "", kernelIDSpec)
	flags.BoolVar(&ebs, "ebs", false, ebsSpec)
	flags.StringVar(&key, "key", "", keySpec)
	flags.StringVar(&tags, "tags", "", tagsSpec)

	clustersNewSpotFleetCmd.MarkFlagRequired("subnets")
	clustersNewSpotFleetCmd.MarkFlagRequired("spot-price")
	clustersNewSpotFleetCmd.MarkFlagRequired("target-capacity")
	clustersNewSpotFleetCmd.MarkFlagRequired("instance-types")
	clustersNewSpotFleetCmd.MarkFlagRequired("security-groups")
}

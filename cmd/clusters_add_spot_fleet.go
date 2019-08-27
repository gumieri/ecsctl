package cmd

import (
	"bytes"
	"encoding/base64"
	"errors"
	"html/template"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/spf13/cobra"
)

func clustersAddSpotFleetRun(cmd *cobra.Command, clusters []string) {
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

	var userData string
	switch platform {
	case "windows", "windows-2016", "windows-2019":
		userData = windowsUserData
	default:
		userData = linuxSpotFleetUserData
	}

	tmpl, err := template.New("UserData").Parse(userData)
	typist.Must(err)

	userDataF := new(bytes.Buffer)
	typist.Must(tmpl.Execute(userDataF, templateUserData{
		Cluster: *c.ClusterName,
		Region:  aws.StringValue(awsSession.Config.Region),
	}))

	if amiID == "" {
		latestAMI, err := latestAmiEcsOptimized(platform)
		typist.Must(err)

		amiID = aws.StringValue(latestAMI.ImageId)
	}

	// TODO: automaticaly --create-roles if does not exist
	spotFleetRoleResponse, err := iamI.GetRole(&iam.GetRoleInput{
		RoleName: aws.String(spotFleetRole),
	})
	typist.Must(err)

	// TODO: automaticaly --create-roles if does not exist
	instanceProfileResponse, err := iamI.GetInstanceProfile(&iam.GetInstanceProfileInput{
		InstanceProfileName: aws.String(instanceProfile),
	})
	typist.Must(err)

	var SecurityGroups []*ec2.GroupIdentifier
	for _, securityGroup := range securityGroups {
		sg, err := findSecurityGroup(securityGroup)
		typist.Must(err)

		SecurityGroups = append(SecurityGroups, &ec2.GroupIdentifier{
			GroupId: sg.GroupId,
		})
	}

	var subnetsIds []string
	for _, subnet := range subnets {
		Subnet, err := findSubnet(subnet)
		typist.Must(err)
		subnetsIds = append(subnetsIds, aws.StringValue(Subnet.SubnetId))
	}

	var LaunchSpecifications []*ec2.SpotFleetLaunchSpecification
	for _, instanceTypeAndWeight := range instanceTypes {
		iTWSlice := strings.Split(instanceTypeAndWeight, ":")
		instanceType := iTWSlice[0]

		var weight float64
		if len(iTWSlice) > 1 {
			weight, err = strconv.ParseFloat(iTWSlice[1], 64)
			typist.Must(err)
		}

		SpotFleetLaunchSpecification := ec2.SpotFleetLaunchSpecification{
			IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
				Arn: instanceProfileResponse.InstanceProfile.Arn,
			},
			EbsOptimized:   aws.Bool(ebs),
			ImageId:        aws.String(amiID),
			InstanceType:   aws.String(instanceType),
			SecurityGroups: SecurityGroups,
			SubnetId:       aws.String(strings.Join(subnetsIds, ",")),
			UserData:       aws.String(base64.StdEncoding.EncodeToString(userDataF.Bytes())),
		}

		if weight > 1 {
			SpotFleetLaunchSpecification.WeightedCapacity = aws.Float64(weight)
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

		if len(tags) > 0 {
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
		TargetCapacity:       aws.Int64(targetCapacity),
	}

	if spotPrice != "" {
		SpotFleetRequestConfig.SpotPrice = aws.String(spotPrice)
	}

	if allocationStrategy != "" {
		SpotFleetRequestConfig.AllocationStrategy = aws.String(allocationStrategy)
	}

	_, err = ec2I.RequestSpotFleet(&ec2.RequestSpotFleetInput{
		SpotFleetRequestConfig: &SpotFleetRequestConfig,
	})
	typist.Must(err)
}

var clustersAddSpotFleetCmd = &cobra.Command{
	Use:     "add-spot-fleet [cluster]",
	Short:   "Add a new Spot Fleet to informed cluster",
	Args:    cobra.ExactArgs(1),
	Aliases: []string{"add-spotfleet"},
	Run:     clustersAddSpotFleetRun,
}

func init() {
	clustersCmd.AddCommand(clustersAddSpotFleetCmd)

	flags := clustersAddSpotFleetCmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVarP(&subnets, "subnet", "n", []string{}, requiredSpec+subnetsSpec)
	flags.StringSliceVarP(&instanceTypes, "instance-type", "i", []string{}, requiredSpec+instanceTypesSpec)
	flags.StringSliceVarP(&securityGroups, "security-group", "g", []string{}, requiredSpec+securityGroupsSpec)
	flags.Int64VarP(&targetCapacity, "target-capacity", "c", 1, targetCapacitySpec)
	flags.StringVar(&instanceProfile, "instance-profile", "ecsInstanceRole", instanceProfileSpec)
	flags.StringVar(&spotFleetRole, "spot-fleet-role", "ecsSpotFleetRole", spotFleetRoleSpec)
	flags.StringVarP(&allocationStrategy, "allocation-strategy", "s", "", allocationStrategySpec)
	flags.StringVar(&platform, "platform", "linux", platformSpec)
	flags.StringVar(&spotPrice, "spot-price", "", spotPriceSpec)
	flags.BoolVar(&monitoring, "monitoring", false, monitoringSpec)
	flags.StringVar(&kernelID, "kernel-id", "", kernelIDSpec)
	flags.StringVar(&amiID, "ami-id", "", amiIDSpec)
	flags.BoolVar(&ebs, "ebs", false, ebsSpec)
	flags.StringVarP(&key, "key", "k", "", keySpec)
	flags.StringSliceVarP(&tags, "tag", "t", []string{}, tagsSpec)

	clustersAddSpotFleetCmd.MarkFlagRequired("subnets")
	clustersAddSpotFleetCmd.MarkFlagRequired("instance-types")
	clustersAddSpotFleetCmd.MarkFlagRequired("security-groups")
}

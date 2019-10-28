package cmd

import (
	"bytes"
	"encoding/base64"
	"errors"
	"html/template"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/spf13/cobra"
)

func clustersAddInstanceRun(cmd *cobra.Command, clusters []string) {
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

	var userData string
	switch platform {
	case "windows", "windows-2016", "windows-2019":
		userData = windowsUserData
	default:
		userData = linuxUserData
	}

	tmpl, err := template.New("UserData").Parse(userData)
	t.Must(err)

	userDataF := new(bytes.Buffer)
	t.Must(tmpl.Execute(userDataF, templateUserData{Cluster: *c.ClusterName}))

	if amiID == "" {
		latestAMI, err := latestAmiEcsOptimized(platform)
		t.Must(err)

		amiID = aws.StringValue(latestAMI.ImageId)
	}

	// TODO: automaticaly --create-roles if does not exist
	if instanceProfile == "" {
		instanceProfile = "ecsInstanceRole"
	}

	instanceProfileResponse, err := iamI.GetInstanceProfile(&iam.GetInstanceProfileInput{
		InstanceProfileName: aws.String(instanceProfile),
	})
	t.Must(err)

	subnetDescription, err := findSubnet(subnet)
	t.Must(err)

	// TODO: AWS Tags
	RunInstancesInput := ec2.RunInstancesInput{
		IamInstanceProfile: &ec2.IamInstanceProfileSpecification{Arn: instanceProfileResponse.InstanceProfile.Arn},
		EbsOptimized:       aws.Bool(ebs),
		ImageId:            aws.String(amiID),
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
		t.Must(err)
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
	t.Must(err)
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

	flags.StringVar(&instanceProfile, "instance-profile", "ecsInstanceRole", instanceProfileSpec)
	flags.StringVar(&platform, "platform", "linux", platformSpec)

	flags.StringVarP(&key, "key", "k", "", keySpec)
	flags.StringSliceVarP(&tags, "tag", "t", []string{}, tagsSpec)
	flags.Int64Var(&minimum, "min", 1, minimumSpec)
	flags.Int64Var(&maximum, "max", 1, maximumSpec)
	flags.StringVar(&credit, "credit", "", creditSpec)
	flags.StringVar(&amiID, "ami-id", "", amiIDSpec)

	clustersAddSpotFleetCmd.MarkFlagRequired("subnet")
	clustersAddSpotFleetCmd.MarkFlagRequired("instance-type")
}

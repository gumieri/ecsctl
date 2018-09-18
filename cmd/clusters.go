package cmd

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/cobra"
)

type templateUserData struct {
	Cluster string
	Region  string
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

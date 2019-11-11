package cmd

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func servicesListRun(cmd *cobra.Command, services []string) {
	clustersDescription, err := ecsI.DescribeClusters(&ecs.DescribeClustersInput{
		Clusters: []*string{
			aws.String(viper.GetString("cluster")),
		},
	})

	t.Must(err)

	if len(clustersDescription.Clusters) == 0 {
		t.Exitln("Source Cluster informed not found")
	}

	c := clustersDescription.Clusters[0]

	servicesList, err := ecsI.ListServices(&ecs.ListServicesInput{
		Cluster: c.ClusterName,
	})

	t.Must(err)

	for _, serviceARN := range servicesList.ServiceArns {
		if listARN {
			t.Outln(aws.StringValue(serviceARN))
			continue
		}

		parsedARN, err := arn.Parse(aws.StringValue(serviceARN))
		t.Must(err)
		parsedResource := strings.Split(parsedARN.Resource, "/")
		t.Outln(parsedResource[len(parsedResource)-1])
	}
}

var servicesListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List services of specified cluster",
	Aliases: []string{"l"},
	Run:     servicesListRun,
}

func init() {
	servicesCmd.AddCommand(servicesListCmd)

	flags := servicesListCmd.Flags()

	flags.StringP("cluster", "c", "", requiredSpec+clusterSpec)
	viper.BindPFlag("cluster", servicesListCmd.Flags().Lookup("cluster"))

	flags.BoolVar(&listARN, "arn", false, listARNSpec)

	servicesListCmd.MarkFlagRequired("cluster")

}

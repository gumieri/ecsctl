package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func servicesListRun(cmd *cobra.Command, services []string) {
	clustersDescription, err := ecsI.DescribeClusters(&ecs.DescribeClustersInput{
		Clusters: []*string{
			aws.String(cluster),
		},
	})

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if len(clustersDescription.Clusters) == 0 {
		fmt.Println(errors.New("Source Cluster informed not found"))
		os.Exit(1)
	}

	c := clustersDescription.Clusters[0]

	servicesList, err := ecsI.ListServices(&ecs.ListServicesInput{
		Cluster: c.ClusterName,
	})

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	for _, service := range servicesList.ServiceArns {
		fmt.Println(aws.StringValue(service))
	}
}

var servicesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List services of specified cluster",
	Run:   servicesListRun,
}

func init() {
	servicesCmd.AddCommand(servicesListCmd)

	flags := servicesListCmd.Flags()

	flags.StringVarP(&cluster, "cluster", "c", "", requiredSpec+clusterSpec)

	servicesCopyCmd.MarkFlagRequired("cluster")

	viper.BindPFlag("cluster", servicesListCmd.Flags().Lookup("cluster"))
}

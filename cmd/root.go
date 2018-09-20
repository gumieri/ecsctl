package cmd

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/iam"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var ecsI *ecs.ECS
var ec2I *ec2.EC2
var iamI *iam.IAM
var cwlI *cloudwatchlogs.CloudWatchLogs

var awsSession *session.Session

func must(err error) {
	if err == nil {
		return
	}

	fmt.Println(err.Error())
	os.Exit(1)
}

func instanceAWSObjects(cmd *cobra.Command, args []string) {
	awsConfig := aws.Config{}

	if region != "" {
		awsConfig.Region = aws.String(region)
	}

	if profile != "" {
		awsConfig.Credentials = credentials.NewSharedCredentials("", profile)
	}

	awsSession = session.New(&awsConfig)

	ecsI = ecs.New(awsSession)
	ec2I = ec2.New(awsSession)
	iamI = iam.New(awsSession)
	cwlI = cloudwatchlogs.New(awsSession)
}

var rootCmd = &cobra.Command{
	Use:              "ecsctl",
	Short:            "Collection of extra functions for AWS ECS",
	PersistentPreRun: instanceAWSObjects,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", cfgFileSpec)

	rootCmd.PersistentFlags().StringVar(&profile, "profile", "", profileSpec)
	viper.BindPFlag("profile", rootCmd.PersistentFlags().Lookup("profile"))

	rootCmd.PersistentFlags().StringVar(&region, "region", "", regionSpec)
	viper.BindPFlag("region", rootCmd.PersistentFlags().Lookup("region"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(".ecsctl")
	}

	viper.AutomaticEnv()

	viper.ReadInConfig()
}

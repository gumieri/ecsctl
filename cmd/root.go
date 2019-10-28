package cmd

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/gumieri/typist"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var ecsI *ecs.ECS
var ecrI *ecr.ECR
var ec2I *ec2.EC2
var iamI *iam.IAM
var cwlI *cloudwatchlogs.CloudWatchLogs

var t = typist.New(&typist.Config{
	Quiet: quiet,
	In:    os.Stdin,
	Out:   os.Stdout,
})

var awsSession *session.Session

func persistentPreRun(cmd *cobra.Command, args []string) {
	awsConfig := aws.Config{}

	if region != "" {
		awsConfig.Region = aws.String(region)
	}

	if profile != "" {
		awsConfig.Credentials = credentials.NewSharedCredentials("", profile)
	}

	awsSession = session.New(&awsConfig)

	ecsI = ecs.New(awsSession)
	ecrI = ecr.New(awsSession)
	ec2I = ec2.New(awsSession)
	iamI = iam.New(awsSession)
	cwlI = cloudwatchlogs.New(awsSession)
}

var rootCmd = &cobra.Command{
	Use:              "ecsctl",
	Short:            "Collection of extra functions for AWS ECS",
	PersistentPreRun: persistentPreRun,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	t.Must(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", cfgFileSpec)

	rootCmd.PersistentFlags().StringVar(&profile, "profile", "", profileSpec)
	viper.BindPFlag("profile", rootCmd.PersistentFlags().Lookup("profile"))

	rootCmd.PersistentFlags().StringVar(&region, "region", "", regionSpec)
	viper.BindPFlag("region", rootCmd.PersistentFlags().Lookup("region"))

	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, quietSpec)
	viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		t.Must(err)

		viper.AddConfigPath(home)
		viper.SetConfigName(".ecsctl")
	}

	viper.AutomaticEnv()

	viper.ReadInConfig()
}

package cmd

import (
	"path"

	"github.com/adrg/xdg"
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

const command = "ecsctl"

var ecsI *ecs.ECS
var ecrI *ecr.ECR
var ec2I *ec2.EC2
var iamI *iam.IAM
var cwlI *cloudwatchlogs.CloudWatchLogs

var t = typist.New(&typist.Config{Quiet: quiet})

var awsSession *session.Session

func persistentPreRun(cmd *cobra.Command, args []string) {
	awsConfig := aws.Config{}

	if r := viper.GetString("region"); r != "" {
		awsConfig.Region = aws.String(r)
	}

	if p := viper.GetString("profile"); p != "" {
		awsConfig.Credentials = credentials.NewSharedCredentials("", p)
	}

	awsSession = session.New(&awsConfig)

	ecsI = ecs.New(awsSession)
	ecrI = ecr.New(awsSession)
	ec2I = ec2.New(awsSession)
	iamI = iam.New(awsSession)
	cwlI = cloudwatchlogs.New(awsSession)
}

var rootCmd = &cobra.Command{
	Use:              command,
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

	rootCmd.PersistentFlags().String("config", "", cfgFileSpec)
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))

	rootCmd.PersistentFlags().String("profile", "", profileSpec)
	viper.BindPFlag("profile", rootCmd.PersistentFlags().Lookup("profile"))

	rootCmd.PersistentFlags().String("region", "", regionSpec)
	viper.BindPFlag("region", rootCmd.PersistentFlags().Lookup("region"))

	rootCmd.PersistentFlags().BoolP("quiet", "q", false, quietSpec)
	viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
}

func initConfig() {
	config := viper.GetString("config")
	if config == "" {
		home, err := homedir.Dir()
		t.Must(err)

		viper.SetConfigName("config")
		viper.AddConfigPath(path.Join(home, "."+command))
		viper.AddConfigPath(path.Join(xdg.ConfigHome, command))
	} else {
		viper.SetConfigFile(config)
	}

	viper.AutomaticEnv()

	viper.ReadInConfig()
}

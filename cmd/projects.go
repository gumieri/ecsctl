package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func projectsRun(cmd *cobra.Command, args []string) {
	cmd.Help()
}

var projectsCmd = &cobra.Command{
	Use:     "projects [command]",
	Short:   "Commands to manage projects at ECS",
	Aliases: []string{"project", "p"},
	Run:     projectsRun,
}

type projectConfig struct {
	Name    string `yaml:"name"`
	Cluster string `yaml:"cluster"`
	Service string `yaml:"service,omitempty"`
}

func recursiveFindConfigFile() (path string, err error) {
	viper.SetConfigName(".ecs-project")

	path, err = os.Getwd()
	if err != nil {
		return
	}

	for {
		viper.AddConfigPath(path)

		err = viper.ReadInConfig()
		if err == nil {
			projectPath = path
			return
		}

		path = filepath.Dir(path)
		if path == string(os.PathSeparator) {
			return
		}
	}
}

func init() {
	recursiveFindConfigFile()

	rootCmd.AddCommand(projectsCmd)
	flags := projectsCmd.PersistentFlags()
	flags.StringVarP(&projectName, "name", "n", viper.GetString("name"), projectNameSpec)
	flags.StringVarP(&projectCluster, "cluster", "c", viper.GetString("cluster"), clusterSpec)
	flags.StringVar(&service, "service", viper.GetString("service"), serviceSpec)
}

package cmd

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/docker/cli/cli/command/image/build"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/idtools"
	"github.com/spf13/cobra"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

func tagNameFromGitHEAD(gitPath string) (tagName string, err error) {
	repo, err := git.PlainOpen(gitPath)
	if err != nil {
		return
	}

	head, err := repo.Head()
	if err != nil {
		return
	}

	tags, err := repo.Tags()
	if err != nil {
		return
	}

	err = tags.ForEach(func(ref *plumbing.Reference) error {
		if ref.Hash() == head.Hash() {
			tagName = ref.Name().Short()
			return nil
		}

		tag, err := repo.TagObject(ref.Hash())
		if err != nil {
			return nil
		}

		if tag.Target == head.Hash() {
			tagName = ref.Name().Short()
		}

		return nil
	})

	return
}

func dockerBuild(ctx context.Context, dockerClient client.Client, path string, tags []string) (output io.ReadCloser, err error) {
	excludes, err := build.ReadDockerignore(path)
	if err != nil {
		return
	}

	dockerfilePath := "." + string(filepath.Separator) + "Dockerfile"
	excludes = build.TrimBuildFilesFromExcludes(excludes, dockerfilePath, false)

	buildCtx, err := archive.TarWithOptions(path, &archive.TarOptions{
		ExcludePatterns: excludes,
		ChownOpts:       &idtools.Identity{UID: 0, GID: 0},
	})
	if err != nil {
		return
	}

	response, err := dockerClient.ImageBuild(ctx, buildCtx, types.ImageBuildOptions{
		Tags:       tags,
		Dockerfile: dockerfilePath,
	})
	if err != nil {
		return
	}

	output = response.Body

	return
}

func ecrURICreateIfNecessary(name string) (repository *ecr.Repository, err error) {
	pName := aws.String(name)

	result, err := ecrI.DescribeRepositories(&ecr.DescribeRepositoriesInput{RepositoryNames: []*string{pName}})
	if err == nil {
		repository = result.Repositories[0]
		return
	}

	aerr, ok := err.(awserr.Error)
	if !ok {
		return
	}

	if aerr.Code() != ecr.ErrCodeRepositoryNotFoundException {
		return
	}

	var response *ecr.CreateRepositoryOutput

	response, err = ecrI.CreateRepository(&ecr.CreateRepositoryInput{RepositoryName: pName})
	if err != nil {
		return
	}

	repository = response.Repository

	return
}

type dockerOutput struct {
	ID       string `json:"id"`
	Status   string `json:"status"`
	Stream   string `json:"stream"`
	Progress string `json:"progress"`
}

func printDocker(output io.ReadCloser) {
	var lastID string

	rd := bufio.NewReader(output)
	for {
		str, err := rd.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return
			}
			continue
		}

		var data dockerOutput
		json.Unmarshal([]byte(str), &data)

		if data.Stream != "" {
			typist.Printf(data.Stream)
		}

		if data.Status != "" {
			if data.ID != lastID {
				typist.Printf("%s - %s", data.ID, data.Status)
			}
		}

		if data.Progress != "" {
			typist.Printf("\r%s", data.Progress)
		}

		lastID = data.ID
	}
}

func projectsDeployRun(cmd *cobra.Command, args []string) {
	ecrName := projectName

	repository, err := ecrURICreateIfNecessary(ecrName)
	typist.Must(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	gitTag, err := tagNameFromGitHEAD(projectPath)
	typist.Must(err)

	dockerTag := aws.StringValue(repository.RepositoryUri)
	if gitTag != "" {
		dockerTag = dockerTag + ":" + gitTag
	}

	dockerClient, err := client.NewClientWithOpts(client.WithVersion("1.37"))
	if err != nil {
		return
	}

	typist.Printf("Build project image: %s\n", dockerTag)
	buildOut, err := dockerBuild(ctx, *dockerClient, projectPath, []string{dockerTag})
	typist.Must(err)
	defer buildOut.Close()
	printDocker(buildOut)

	authResponse, err := ecrI.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{
		RegistryIds: []*string{repository.RegistryId},
	})
	typist.Must(err)

	tokenS, err := base64.StdEncoding.DecodeString(aws.StringValue(authResponse.AuthorizationData[0].AuthorizationToken))
	typist.Must(err)

	tokenData := strings.Split(string(tokenS), ":")
	authConfig := types.AuthConfig{
		Username:      tokenData[0],
		Password:      tokenData[1],
		ServerAddress: aws.StringValue(authResponse.AuthorizationData[0].ProxyEndpoint),
	}
	// typist.Must(dockerClient.RegistryLogin(ctx, authConfig)) // TODO: check if necessary
	authBytes, _ := json.Marshal(authConfig)

	pushOut, err := dockerClient.ImagePush(ctx, dockerTag, types.ImagePushOptions{
		RegistryAuth: base64.URLEncoding.EncodeToString(authBytes),
	})

	typist.Must(err)
	defer pushOut.Close()
	printDocker(pushOut)

	// cluster = projectCluster

	// clustersDescription, err := ecsI.DescribeClusters(&ecs.DescribeClustersInput{
	// 	Clusters: []*string{
	// 		aws.String(cluster),
	// 	},
	// })

	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	os.Exit(1)
	// }

	// if len(clustersDescription.Clusters) == 0 {
	// 	fmt.Println(errors.New("Cluster informed not found"))
	// 	os.Exit(1)
	// }

	// c := clustersDescription.Clusters[0]

	// servicesDescription, err := ecsI.DescribeServices(&ecs.DescribeServicesInput{
	// 	Cluster: c.ClusterName,
	// 	Services: []*string{
	// 		aws.String(service),
	// 	},
	// })

	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	os.Exit(1)
	// }

	// if len(servicesDescription.Services) == 0 {
	// 	fmt.Println(errors.New("Service informed not found"))
	// 	os.Exit(1)
	// }

	// s := servicesDescription.Services[0]

	// tdDescription, err := ecsI.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
	// 	TaskDefinition: s.TaskDefinition,
	// })

	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	os.Exit(1)
	// }

	// td := tdDescription.TaskDefinition

	// var cdToUpdate *ecs.ContainerDefinition

	// if containerName == "" {
	// 	cdToUpdate = td.ContainerDefinitions[0]
	// } else {
	// 	for _, cd := range td.ContainerDefinitions {
	// 		if aws.StringValue(cd.Name) == containerName {
	// 			cdToUpdate = cd
	// 			break
	// 		}
	// 	}
	// }

	// if cdToUpdate == nil {
	// 	fmt.Println(fmt.Errorf("No container on the Task Family %s", aws.StringValue(td.Family)))
	// 	os.Exit(1)
	// }

	// if tag != "" {
	// 	image = strings.Split(aws.StringValue(cdToUpdate.Image), ":")[0] + ":" + tag
	// }

	// cdToUpdate.Image = aws.String(image)

	// newTDDescription, err := ecsI.RegisterTaskDefinition(&ecs.RegisterTaskDefinitionInput{
	// 	ContainerDefinitions:    td.ContainerDefinitions,
	// 	Cpu:                     td.Cpu,
	// 	ExecutionRoleArn:        td.ExecutionRoleArn,
	// 	Family:                  td.Family,
	// 	Memory:                  td.Memory,
	// 	NetworkMode:             td.NetworkMode,
	// 	PlacementConstraints:    td.PlacementConstraints,
	// 	RequiresCompatibilities: td.RequiresCompatibilities,
	// 	TaskRoleArn:             td.TaskRoleArn,
	// 	Volumes:                 td.Volumes,
	// })

	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	os.Exit(1)
	// }

	// newTD := newTDDescription.TaskDefinition
	// oldFamilyRevision := aws.StringValue(td.Family) + ":" + strconv.FormatInt(aws.Int64Value(td.Revision), 10)

	// _, err = ecsI.DeregisterTaskDefinition(&ecs.DeregisterTaskDefinitionInput{
	// 	TaskDefinition: aws.String(oldFamilyRevision),
	// })

	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	os.Exit(1)
	// }

	// newFamilyRevision := aws.StringValue(newTD.Family) + ":" + strconv.FormatInt(aws.Int64Value(newTD.Revision), 10)

	// _, err = ecsI.UpdateService(&ecs.UpdateServiceInput{
	// 	Cluster:        c.ClusterName,
	// 	Service:        aws.String(service),
	// 	TaskDefinition: aws.String(newFamilyRevision),
	// })

	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	os.Exit(1)
	// }
}

var projectsDeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy a project",
	Run:   projectsDeployRun,
}

func init() {
	projectsCmd.AddCommand(projectsDeployCmd)
	flags := projectsDeployCmd.Flags()
	flags.StringVar(&containerName, "container", "", containerNameSpec)
	flags.StringVarP(&tag, "tag", "t", "", tagSpec)
	flags.StringVarP(&image, "image", "i", "", imageSpec)
	flags.StringVarP(&repository, "repository", "r", "", repositorySpec)
}

package cmd

import (
	"encoding/json"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/TylerBrock/colorjson"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type outputConfiguration struct {
	Expand         bool
	Raw            bool
	RawString      bool
	HideStreamName bool
	HideDate       bool
	Invert         bool
	NoColor        bool
}

func (c *outputConfiguration) Formatter() *colorjson.Formatter {
	formatter := colorjson.NewFormatter()

	if c.Expand {
		formatter.Indent = 4
	}

	if c.RawString {
		formatter.RawStrings = true
	}

	if c.Invert {
		formatter.KeyColor = color.New(color.FgBlack)
	}

	if c.NoColor {
		color.NoColor = true
	}

	return formatter
}

func printEvent(formatter *colorjson.Formatter, event *cloudwatchlogs.FilteredLogEvent) {
	red := color.New(color.FgRed).SprintFunc()
	white := color.New(color.FgWhite).SprintFunc()

	str := aws.StringValue(event.Message)
	bytes := []byte(str)
	date := aws.MillisecondsTimeValue(event.Timestamp)
	dateStr := date.Format(time.RFC3339)
	streamStr := aws.StringValue(event.LogStreamName)
	jl := map[string]interface{}{}

	err := json.Unmarshal(bytes, &jl)
	if err == nil {
		output, _ := formatter.Marshal(jl)
		t.Outf("[%s] (%s) %s\n", red(dateStr), white(streamStr), output)
		return
	}

	t.Outf("[%s] (%s) %s\n", red(dateStr), white(streamStr), str)
}

func taskDefinitionsRunRun(cmd *cobra.Command, args []string) {
	cluster := viper.GetString("cluster")

	taskDefinitionFamily := args[0]

	tdDescription, err := ecsI.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(taskDefinitionFamily),
	})

	t.Must(err)

	td := tdDescription.TaskDefinition

	if revision == "" {
		revision = strconv.FormatInt(aws.Int64Value(td.Revision), 10)
	}

	runTaskInput := &ecs.RunTaskInput{
		Cluster:        aws.String(cluster),
		TaskDefinition: aws.String(taskDefinitionFamily + ":" + revision),
		StartedBy:      aws.String("ecsctl"),
		Count:          aws.Int64(numberTasks),
	}

	if groupTasks != "" {
		runTaskInput.Group = aws.String(groupTasks)
	}

	envOverride := make([]*ecs.KeyValuePair, 0)
	for _, envPair := range environmentVariables {
		envSlice := strings.Split(envPair, "=")
		if len(envSlice) != 2 {
			continue
		}

		envOverride = append(envOverride, &ecs.KeyValuePair{Name: aws.String(envSlice[0]), Value: aws.String(envSlice[1])})
	}

	if len(envOverride) > 0 {
		runTaskInput.Overrides = &ecs.TaskOverride{
			ContainerOverrides: []*ecs.ContainerOverride{{
				Name:        td.ContainerDefinitions[0].Name,
				Environment: envOverride,
			}},
		}
	}

	if memory > 0 {
		if runTaskInput.Overrides == nil {
			runTaskInput.Overrides = &ecs.TaskOverride{
				ContainerOverrides: []*ecs.ContainerOverride{{
					Name:   td.ContainerDefinitions[0].Name,
					Memory: aws.Int64(memory),
				}},
			}
		} else {
			runTaskInput.Overrides.ContainerOverrides[0].Memory = aws.Int64(memory)
		}
	}

	if memoryReservation > 0 {
		if runTaskInput.Overrides == nil {
			runTaskInput.Overrides = &ecs.TaskOverride{
				ContainerOverrides: []*ecs.ContainerOverride{{
					Name:              td.ContainerDefinitions[0].Name,
					MemoryReservation: aws.Int64(memoryReservation),
				}},
			}
		} else {
			runTaskInput.Overrides.ContainerOverrides[0].MemoryReservation = aws.Int64(memoryReservation)
		}
	}

	if commandOverride != "" {
		commandSlice := aws.StringSlice(strings.Split(commandOverride, " "))

		if runTaskInput.Overrides == nil {
			runTaskInput.Overrides = &ecs.TaskOverride{
				ContainerOverrides: []*ecs.ContainerOverride{{
					Name:    td.ContainerDefinitions[0].Name,
					Command: commandSlice,
				}},
			}
		} else {
			runTaskInput.Overrides.ContainerOverrides[0].Command = commandSlice
		}
	}

	retryCount := 0
	retryLimit := 5
	var taskResult *ecs.RunTaskOutput
	for {
		taskResult, err = ecsI.RunTask(runTaskInput)

		if err != nil || retryCount >= retryLimit || len(taskResult.Tasks) > 0 {
			break
		}

		retryCount = retryCount + 1

		time.Sleep(1 * time.Second)
	}

	t.Must(err)

	if len(taskResult.Tasks) == 0 {
		t.Exitln("task failed to run")
	}

	tSplited := strings.Split(aws.StringValue(taskResult.Tasks[0].TaskArn), "/")
	taskID := tSplited[2]

	if viper.GetBool("output-ip") {
		var tasksStatus *ecs.DescribeTasksOutput
		var err error

		for {
			tasksStatus, err = ecsI.DescribeTasks(&ecs.DescribeTasksInput{
				Cluster: aws.String(cluster),
				Tasks:   []*string{aws.String(taskID)},
			})
			t.Must(err)

			status := aws.StringValue(tasksStatus.Tasks[0].LastStatus)
			if status != "PENDING" {
				break
			}

			time.Sleep(1 * time.Second)
		}

		ec2ECSInstance, err := ecsI.DescribeContainerInstances(&ecs.DescribeContainerInstancesInput{
			Cluster:            aws.String(cluster),
			ContainerInstances: []*string{tasksStatus.Tasks[0].ContainerInstanceArn},
		})
		t.Must(err)

		if len(ec2ECSInstance.ContainerInstances) == 0 {
			t.Exitln("failed to find the EC2 Instance running this task")
		}

		ec2Info, err := ec2I.DescribeInstances(&ec2.DescribeInstancesInput{
			InstanceIds: []*string{ec2ECSInstance.ContainerInstances[0].Ec2InstanceId},
		})

		if len(ec2Info.Reservations) == 0 || len(ec2Info.Reservations[0].Instances) == 0 {
			t.Exitln("failed to describe the EC2 Instance running this task")
		}

		hostIp := ec2Info.Reservations[0].Instances[0].PublicIpAddress
		if hostIp == nil {
			hostIp = ec2Info.Reservations[0].Instances[0].PrivateIpAddress
		}

		hostPort := tasksStatus.Tasks[0].Containers[0].NetworkBindings[0].HostPort

		t.Outf("%s:%d\n", aws.StringValue(hostIp), aws.Int64Value(hostPort))
	}

	if !follow {
		t.Exit(nil)
	}

	if exit {
		var gracefulStop = make(chan os.Signal, 1)
		signal.Notify(gracefulStop, syscall.SIGTERM, syscall.SIGINT)
		go func() {
			<-gracefulStop

			ecsI.StopTask(&ecs.StopTaskInput{
				Cluster: aws.String(cluster),
				Task:    taskResult.Tasks[0].TaskArn,
			})

			t.Exit(nil)
		}()
	}

	logDriver := td.ContainerDefinitions[0].LogConfiguration.LogDriver
	if aws.StringValue(logDriver) != "awslogs" {
		t.Exit(nil)
	}

	logPrefix := td.ContainerDefinitions[0].LogConfiguration.Options["awslogs-stream-prefix"]
	logGroup := td.ContainerDefinitions[0].LogConfiguration.Options["awslogs-group"]

	cName := td.ContainerDefinitions[0].Name
	logStreamName := aws.StringValue(logPrefix) + "/" + aws.StringValue(cName) + "/" + taskID

	var lastSeenTime *int64
	var seenEventIDs map[string]bool
	output := outputConfiguration{}
	formatter := output.Formatter()

	clearSeenEventIds := func() {
		seenEventIDs = make(map[string]bool, 0)
	}

	addSeenEventIDs := func(id *string) {
		seenEventIDs[*id] = true
	}

	updateLastSeenTime := func(ts *int64) {
		if lastSeenTime == nil || *ts > *lastSeenTime {
			lastSeenTime = ts
			clearSeenEventIds()
		}
	}

	cwInput := cloudwatchlogs.FilterLogEventsInput{
		LogGroupName:   logGroup,
		LogStreamNames: []*string{aws.String(logStreamName)},
	}

	handlePage := func(page *cloudwatchlogs.FilterLogEventsOutput, lastPage bool) bool {
		for _, event := range page.Events {
			updateLastSeenTime(event.Timestamp)
			if _, seen := seenEventIDs[*event.EventId]; !seen {
				printEvent(formatter, event)
				addSeenEventIDs(event.EventId)
			}
		}
		return !lastPage
	}

	retryCount = 0
	retryLimit = 50
	for {
		err := cwlI.FilterLogEventsPages(&cwInput, handlePage)
		if err != nil {
			retryCount = retryCount + 1

			if retryCount >= retryLimit {
				t.Exit(err)
			}
		}

		if lastSeenTime != nil {
			cwInput.SetStartTime(*lastSeenTime)
		}

		tasksStatus, err := ecsI.DescribeTasks(&ecs.DescribeTasksInput{
			Cluster: aws.String(cluster),
			Tasks:   []*string{aws.String(taskID)},
		})

		if err != nil {
			t.Exit(err)
		}

		status := aws.StringValue(tasksStatus.Tasks[0].LastStatus)
		if status == "STOPPED" {
			os.Exit(int(aws.Int64Value(tasksStatus.Tasks[0].Containers[0].ExitCode)))
		}

		time.Sleep(5 * time.Second)
	}
}

var taskDefinitionsRunCmd = &cobra.Command{
	Use:              "run [task-definition]",
	Short:            "Run a Task Definition",
	Args:             cobra.ExactArgs(1),
	Run:              taskDefinitionsRunRun,
	PersistentPreRun: persistentPreRun,
}

func init() {
	taskDefinitionsCmd.AddCommand(taskDefinitionsRunCmd)

	flags := taskDefinitionsRunCmd.Flags()

	flags.BoolVar(&exit, "exit", false, exitSpec)

	flags.BoolVarP(&follow, "follow", "f", false, followSpec)

	flags.StringVar(&revision, "revision", "", revisionSpec)

	flags.Int64VarP(&numberTasks, "number", "n", 1, numberTasksSpec)

	flags.StringSliceVarP(&environmentVariables, "env", "e", []string{}, environmentVariablesSpec)

	flags.StringVar(&commandOverride, "command", "", "Command to override in the (container definition) task execution")

	flags.Int64Var(&memory, "memory", 0, "Memory in MiB (Hard limit) to override in (container definition) task execution")

	flags.Int64Var(&memoryReservation, "memory-reservation", 0, "Memory reservation in MiB (soft limit) to override in (container definition) task execution")

	flags.StringVar(&groupTasks, "group", "", groupTasksSpec)

	flags.Bool("output-ip", false, "Return the Private IP and Port of the Container on running Task for EC2 Instances")
	viper.BindPFlag("output-ip", taskDefinitionsRunCmd.Flags().Lookup("output-ip"))

	flags.StringP("cluster", "c", "default", clusterSpec)
	viper.BindPFlag("cluster", taskDefinitionsRunCmd.Flags().Lookup("cluster"))
}

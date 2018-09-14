package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	editor "github.com/gumieri/open-in-editor"
	"github.com/spf13/cobra"
)

var editorCommand string

func taskDefinitionsEditRun(cmd *cobra.Command, args []string) {
	taskDefinition := args[0]

	tdDescription, err := ecsI.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(taskDefinition),
	})

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	td := tdDescription.TaskDefinition

	newTD := ecs.RegisterTaskDefinitionInput{
		ContainerDefinitions:    td.ContainerDefinitions,
		Cpu:                     td.Cpu,
		ExecutionRoleArn:        td.ExecutionRoleArn,
		Family:                  td.Family,
		Memory:                  td.Memory,
		NetworkMode:             td.NetworkMode,
		PlacementConstraints:    td.PlacementConstraints,
		RequiresCompatibilities: td.RequiresCompatibilities,
		TaskRoleArn:             td.TaskRoleArn,
		Volumes:                 td.Volumes,
	}

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if editorCommand == "" {
		editorCommand = "vim"
	}

	jsonTdDescription, err := json.MarshalIndent(newTD, "", "  ")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	edited, err := editor.GetContentFromTemporaryFile(editorCommand, taskDefinition+".json", string(jsonTdDescription))
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	var editedTD *ecs.RegisterTaskDefinitionInput
	err = json.Unmarshal([]byte(edited), &editedTD)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	newTDDescription, err := ecsI.RegisterTaskDefinition(editedTD)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	newFamilyRevision := aws.StringValue(newTDDescription.TaskDefinition.Family) + ":" + strconv.FormatInt(aws.Int64Value(newTDDescription.TaskDefinition.Revision), 10)

	fmt.Println(newFamilyRevision)

	oldFamilyRevision := aws.StringValue(td.Family) + ":" + strconv.FormatInt(aws.Int64Value(td.Revision), 10)
	_, err = ecsI.DeregisterTaskDefinition(&ecs.DeregisterTaskDefinitionInput{TaskDefinition: aws.String(oldFamilyRevision)})

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

var taskDefinitionsEditCmd = &cobra.Command{
	Use:   "edit [task-definition]",
	Short: "Edit a Task Definition",
	Args:  cobra.ExactArgs(1),
	Run:   taskDefinitionsEditRun,
}

func init() {
	taskDefinitionsCmd.AddCommand(taskDefinitionsEditCmd)

	taskDefinitionsEditCmd.Flags().StringVar(&editorCommand, "editor", "", "Override default text editor.")
}

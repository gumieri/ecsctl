package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	oie "github.com/gumieri/open-in-editor"
	"github.com/spf13/cobra"
)

func taskDefinitionsEditRun(cmd *cobra.Command, args []string) {
	taskDefinition := args[0]

	if editorCommand == "" {
		editorCommand = os.Getenv("EDITOR")
	}

	if editorCommand == "" {
		typist.Must(errors.New("no editor defined"))
	}

	tdDescription, err := ecsI.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(taskDefinition),
	})
	typist.Must(err)

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

	jsonTdDescription, err := json.MarshalIndent(newTD, "", "  ")
	typist.Must(err)

	editor := oie.Editor{Command: editorCommand}
	typist.Must(editor.OpenTempFile(&oie.File{
		FileName: taskDefinition + ".json",
		Content:  jsonTdDescription,
	}))

	file, err := editor.LastFile()
	typist.Must(err)

	var editedTD *ecs.RegisterTaskDefinitionInput
	typist.Must(json.Unmarshal(file.Content, &editedTD))

	newTDDescription, err := ecsI.RegisterTaskDefinition(editedTD)
	typist.Must(err)

	newFamilyRevision := aws.StringValue(newTDDescription.TaskDefinition.Family) + ":" + strconv.FormatInt(aws.Int64Value(newTDDescription.TaskDefinition.Revision), 10)

	fmt.Println(newFamilyRevision)

	oldFamilyRevision := aws.StringValue(td.Family) + ":" + strconv.FormatInt(aws.Int64Value(td.Revision), 10)
	_, err = ecsI.DeregisterTaskDefinition(&ecs.DeregisterTaskDefinitionInput{
		TaskDefinition: aws.String(oldFamilyRevision),
	})
	typist.Must(err)
}

var taskDefinitionsEditCmd = &cobra.Command{
	Use:   "edit [task-definition]",
	Short: "Edit a Task Definition",
	Args:  cobra.ExactArgs(1),
	Run:   taskDefinitionsEditRun,
}

func init() {
	taskDefinitionsCmd.AddCommand(taskDefinitionsEditCmd)

	taskDefinitionsEditCmd.Flags().StringVar(&editorCommand, "editor", "", editorCommandSpec)
}

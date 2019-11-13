package cmd

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"

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
		t.Must(errors.New("no editor defined"))
	}

	tdDescription, err := ecsI.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(taskDefinition),
	})
	t.Must(err)

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
	t.Must(err)

	editor := oie.Editor{Command: editorCommand}
	t.Must(editor.OpenTempFile(&oie.File{
		FileName: taskDefinition + ".json",
		Content:  jsonTdDescription,
	}))

	file, err := editor.LastFile()
	t.Must(err)

	if strings.Trim(string(file.Content), "\n") == strings.Trim(string(jsonTdDescription), "\n") {
		t.Errorln("nothing changed")
		return
	}

	var editedTD *ecs.RegisterTaskDefinitionInput
	t.Must(json.Unmarshal(file.Content, &editedTD))

	newTDDescription, err := ecsI.RegisterTaskDefinition(editedTD)
	t.Must(err)

	newFamilyRevision := aws.StringValue(newTDDescription.TaskDefinition.Family) + ":" + strconv.FormatInt(aws.Int64Value(newTDDescription.TaskDefinition.Revision), 10)

	t.Outln(newFamilyRevision)

	oldFamilyRevision := aws.StringValue(td.Family) + ":" + strconv.FormatInt(aws.Int64Value(td.Revision), 10)

	t.Must(ecsI.DeregisterTaskDefinition(&ecs.DeregisterTaskDefinitionInput{
		TaskDefinition: aws.String(oldFamilyRevision),
	}))
}

var taskDefinitionsEditCmd = &cobra.Command{
	Use:              "edit [task-definition]",
	Short:            "Edit a Task Definition",
	Args:             cobra.ExactArgs(1),
	Run:              taskDefinitionsEditRun,
	PersistentPreRun: persistentPreRun,
}

func init() {
	taskDefinitionsCmd.AddCommand(taskDefinitionsEditCmd)

	taskDefinitionsEditCmd.Flags().StringVar(&editorCommand, "editor", "", editorCommandSpec)
}

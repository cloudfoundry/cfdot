package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/cfdot/commands/helpers"
	"github.com/openzipkin/zipkin-go/idgenerator"
	"github.com/spf13/cobra"
)

var createTaskCmd = &cobra.Command{
	Use:   "create-task (SPEC|@FILE)",
	Short: "Create a Task",
	Long:  "Create a Task from the given spec. Spec can either be json encoded task, e.g. '{\"task_guid\":\"some-guid\"}' or a file containing json encoded task, e.g. @/path/to/spec/file",
	RunE:  createTask,
}

func init() {
	AddBBSAndTimeoutFlags(createTaskCmd)
	RootCmd.AddCommand(createTaskCmd)
}

func createTask(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return NewCFDotValidationError(cmd, fmt.Errorf("missing spec argument"))
	}

	spec, err := ValidateCreateTaskArguments(args)
	if err != nil {
		return NewCFDotValidationError(cmd, err)
	}

	bbsClient, err := helpers.NewBBSClient(cmd, Config)
	if err != nil {
		return NewCFDotError(cmd, err)
	}

	err = CreateTask(cmd.OutOrStdout(), cmd.OutOrStderr(), bbsClient, spec)
	if err != nil {
		return NewCFDotError(cmd, err)
	}

	return nil
}

func ValidateCreateTaskArguments(args []string) ([]byte, error) {
	var spec []byte
	var err error
	var task *models.Task

	argValue := args[0]
	if strings.HasPrefix(argValue, "@") {
		_, err := os.Stat(argValue[1:])
		if err != nil {
			return nil, err
		}
		spec, err = ioutil.ReadFile(argValue[1:])
		if err != nil {
			return nil, err
		}

	} else {
		spec = []byte(argValue)
	}
	err = json.Unmarshal([]byte(spec), &task)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Invalid JSON: %s", err.Error()))
	}
	return spec, nil
}

func CreateTask(stdout, stderr io.Writer, bbsClient bbs.Client, spec []byte) error {
	logger := globalLogger.Session("create-task")

	var task *models.Task
	err := json.Unmarshal(spec, &task)
	if err != nil {
		return err
	}

	traceID := idgenerator.NewRandom128().TraceID().String()
	err = bbsClient.DesireTask(logger, traceID, task.TaskGuid, task.Domain, task.TaskDefinition)
	if err != nil {
		return err
	}

	return nil
}

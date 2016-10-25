package commands

import (
	"encoding/json"
	"io"

	"code.cloudfoundry.org/bbs"
	"github.com/spf13/cobra"
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Display task",
	Long:  "Retrieve the specified task",
	RunE:  task,
}

func init() {
	AddBBSFlags(taskCmd)
	RootCmd.AddCommand(taskCmd)
}

func task(cmd *cobra.Command, args []string) error {
	guid, err := ValidateTaskArgs(args)
	if err != nil {
		return NewCFDotValidationError(cmd, err)
	}

	bbsClient, err := newBBSClient(cmd)
	if err != nil {
		return NewCFDotError(cmd, err)
	}

	if err := TaskByGuid(cmd.OutOrStdout(), cmd.OutOrStderr(), bbsClient, guid); err != nil {
		return NewCFDotError(cmd, err)
	}

	return nil
}

func TaskByGuid(stdout, _ io.Writer, bbsClient bbs.Client, taskGuid string) error {
	logger := globalLogger.Session("task-by-guid")

	task, err := bbsClient.TaskByGuid(logger, taskGuid)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(stdout)
	err = encoder.Encode(task)
	if err != nil {
		logger.Error("failed-to-marshal", err)
	}

	return nil
}

func ValidateTaskArgs(args []string) (string, error) {
	if len(args) == 0 || args[0] == "" {
		return "", errMissingArguments
	}

	if len(args) > 1 {
		return "", errExtraArguments
	}

	return args[0], nil
}

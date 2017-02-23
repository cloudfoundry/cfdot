package commands

import (
	"io"

	"code.cloudfoundry.org/bbs"
	"github.com/spf13/cobra"
)

var deleteTaskCmd = &cobra.Command{
	Use:   "delete-task TASK_GUID",
	Short: "Delete a Task",
	Long:  "Delete a Task with the given task guid.",
	RunE:  deleteTask,
}

func init() {
	AddBBSFlags(deleteTaskCmd)
	RootCmd.AddCommand(deleteTaskCmd)
}

func deleteTask(cmd *cobra.Command, args []string) error {
	taskGuid, err := ValidateDeleteTaskArguments(args)
	if err != nil {
		return NewCFDotValidationError(cmd, err)
	}

	bbsClient, err := newBBSClient(cmd)
	if err != nil {
		return NewCFDotError(cmd, err)
	}

	err = DeleteTask(cmd.OutOrStdout(), cmd.OutOrStderr(), bbsClient, taskGuid)
	if err != nil {
		return NewCFDotError(cmd, err)
	}

	return nil
}

func ValidateDeleteTaskArguments(args []string) (string, error) {
	if len(args) == 0 {
		return "", errMissingArguments
	}

	if len(args) > 1 {
		return "", errExtraArguments
	}

	if args[0] == "" {
		return "", errInvalidProcessGuid
	}

	return args[0], nil
}

func DeleteTask(stdout, stderr io.Writer, bbsClient bbs.Client, taskGuid string) error {
	logger := globalLogger.Session("delete-task")
	err := bbsClient.ResolvingTask(logger, taskGuid)
	if err != nil {
		return err
	}
	err = bbsClient.DeleteTask(logger, taskGuid)
	if err != nil {
		return err
	}

	return nil
}

package commands

import (
	"encoding/json"
	"io"

	"code.cloudfoundry.org/bbs"
	"github.com/spf13/cobra"
)

var tasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "List tasks in BBS",
	Long:  "List all tasks in BBS",
	RunE:  tasks,
}

func init() {
	AddBBSFlags(tasksCmd)
	RootCmd.AddCommand(tasksCmd)
}

func tasks(cmd *cobra.Command, args []string) error {
	err := ValidateTasksArgs(args)
	if err != nil {
		return NewCFDotValidationError(cmd, err)
	}

	bbsClient, err := newBBSClient(cmd)
	if err != nil {
		return NewCFDotError(cmd, err)
	}

	err = Tasks(cmd.OutOrStdout(), cmd.OutOrStderr(), bbsClient)
	if err != nil {
		return NewCFDotError(cmd, err)
	}

	return nil
}

func Tasks(stdout, _ io.Writer, bbsClient bbs.Client) error {
	tasks, err := bbsClient.Tasks(globalLogger)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(stdout)
	for _, task := range tasks {
		err = encoder.Encode(task)
		if err != nil {
			return err
		}
	}

	return nil
}

func ValidateTasksArgs(args []string) error {
	if len(args) > 0 {
		return errExtraArguments
	}
	return nil
}

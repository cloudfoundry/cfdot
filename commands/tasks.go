package commands

import (
	"encoding/json"
	"io"

	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/bbs/models"
	"github.com/spf13/cobra"
)

var tasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "List tasks in BBS",
	Long:  "List all tasks in BBS",
	RunE:  tasks,
}

// flags
var tasksDomainFlag, tasksCellIdFlag string

func init() {
	AddBBSFlags(tasksCmd)
	tasksCmd.Flags().StringVarP(&tasksDomainFlag, "domain", "d", "", "retrieve only tasks for the given domain")
	tasksCmd.Flags().StringVarP(&tasksCellIdFlag, "cell-id", "c", "", "retrieve only tasks for the given cell-id")
	RootCmd.AddCommand(tasksCmd)
}

func tasks(cmd *cobra.Command, args []string) error {
	err := ValidateConflictingShortAndLongFlag("-d", "--domain", cmd)
	if err != nil {
		return NewCFDotValidationError(cmd, err)
	}

	err = ValidateConflictingShortAndLongFlag("-c", "--cell-id", cmd)
	if err != nil {
		return NewCFDotValidationError(cmd, err)
	}

	err = ValidateTasksArgs(args)
	if err != nil {
		return NewCFDotValidationError(cmd, err)
	}

	bbsClient, err := newBBSClient(cmd)
	if err != nil {
		return NewCFDotError(cmd, err)
	}

	err = Tasks(cmd.OutOrStdout(), cmd.OutOrStderr(), bbsClient, tasksDomainFlag, tasksCellIdFlag)
	if err != nil {
		return NewCFDotError(cmd, err)
	}

	return nil
}

func Tasks(stdout, _ io.Writer, bbsClient bbs.Client, domain, cellID string) error {
	var tasks []*models.Task
	var err error

	tasks, err = bbsClient.TasksWithFilter(globalLogger, models.TaskFilter{Domain: domain, CellID: cellID})
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

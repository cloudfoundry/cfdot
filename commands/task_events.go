package commands

import (
	"encoding/json"
	"io"

	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/bbs/models"
	"github.com/spf13/cobra"
)

var (
	taskEventsCellIdFlag string
)

var taskEventsCmd = &cobra.Command{
	Use:   "task-events",
	Short: "Subscribe to BBS Task events",
	Long:  "Subscribe to BBS Task events",
	RunE:  taskEvents,
}

type TaskEvent struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func init() {
	AddBBSFlags(taskEventsCmd)
	RootCmd.AddCommand(taskEventsCmd)
}

func taskEvents(cmd *cobra.Command, args []string) error {
	err := validateLRPEventsArguments(args)
	if err != nil {
		return NewCFDotValidationError(cmd, err)
	}

	bbsClient, err := newBBSClient(cmd)
	if err != nil {
		return NewCFDotError(cmd, err)
	}

	err = TaskEvents(cmd.OutOrStdout(), cmd.OutOrStderr(), bbsClient, taskEventsCellIdFlag)
	if err != nil {
		return NewCFDotError(cmd, err)
	}
	return nil
}

func validateTaskEventsArguments(args []string) error {
	if len(args) > 0 {
		return errExtraArguments
	}
	return nil
}

func TaskEvents(stdout, stderr io.Writer, bbsClient bbs.Client, cellID string) error {
	logger := globalLogger.Session("lrp-events")

	es, err := bbsClient.SubscribeToTaskEvents(logger)
	if err != nil {
		return models.ConvertError(err)
	}
	defer es.Close()
	encoder := json.NewEncoder(stdout)

	var taskEvents LRPEvent
	for {
		event, err := es.Next()
		switch err {
		case nil:
			taskEvents.Type = event.EventType()
			taskEvents.Data = event
			err = encoder.Encode(taskEvents)
			if err != nil {
				logger.Error("failed-to-marshal", err)
			}
		case io.EOF:
			return nil
		default:
			return err
		}
	}
	return nil
}

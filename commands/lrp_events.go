package commands

import (
	"encoding/json"
	"io"

	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/cfdot/commands/helpers"
	"github.com/spf13/cobra"
)

var (
	lrpEventsCellIdFlag string
)

var lrpEventsCmd = &cobra.Command{
	Use:   "lrp-events",
	Short: "Subscribe to BBS LRP events",
	Long:  "Subscribe to BBS LRP events",
	RunE:  lrpEvents,
}

type LRPEvent struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func init() {
	AddBBSFlags(lrpEventsCmd)

	lrpEventsCmd.Flags().StringVarP(&lrpEventsCellIdFlag, "cell-id", "c", "", "retrieve only events for the given cell id")

	RootCmd.AddCommand(lrpEventsCmd)
}

func lrpEvents(cmd *cobra.Command, args []string) error {
	err := validateLRPEventsArguments(args)
	if err != nil {
		return NewCFDotValidationError(cmd, err)
	}

	bbsClient, err := helpers.NewBBSClient(cmd, Config)
	if err != nil {
		return NewCFDotError(cmd, err)
	}

	err = LRPEvents(cmd.OutOrStdout(), cmd.OutOrStderr(), bbsClient, lrpEventsCellIdFlag)
	if err != nil {
		return NewCFDotError(cmd, err)
	}
	return nil
}

func validateLRPEventsArguments(args []string) error {
	if len(args) > 0 {
		return errExtraArguments
	}
	return nil
}

func LRPEvents(stdout, stderr io.Writer, bbsClient bbs.Client, cellID string) error {
	logger := globalLogger.Session("lrp-events")

	es, err := bbsClient.SubscribeToEventsByCellID(logger, cellID)
	if err != nil {
		return models.ConvertError(err)
	}
	defer es.Close()
	encoder := json.NewEncoder(stdout)

	var lrpEvent LRPEvent
	for {
		event, err := es.Next()
		switch err {
		case nil:
			lrpEvent.Type = event.EventType()
			lrpEvent.Data = event
			err = encoder.Encode(lrpEvent)
			if err != nil {
				logger.Error("failed-to-marshal", err)
			}
		case io.EOF:
			return nil
		default:
			return err
		}
	}
}

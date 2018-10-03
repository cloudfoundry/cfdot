package commands

import (
	"encoding/json"
	"io"

	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/bbs/events"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/cfdot/commands/helpers"
	multierror "github.com/hashicorp/go-multierror"
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

	oldES, err := bbsClient.SubscribeToEventsByCellID(logger, cellID)
	if err != nil {
		return models.ConvertError(err)
	}
	defer oldES.Close()

	instanceES, err := bbsClient.SubscribeToInstanceEventsByCellID(logger, cellID)
	if err != nil {
		return models.ConvertError(err)
	}
	defer instanceES.Close()

	encoder := json.NewEncoder(stdout)

	oldEventStream := make(chan models.Event)
	newEventStream := make(chan models.Event)
	oldErrChan := make(chan error)
	newErrChan := make(chan error)

	readEvent := func(es events.EventSource, eventStreamChan chan models.Event, errChan chan error) {
		for {
			event, err := es.Next()
			if err != nil {
				errChan <- err
				return
			}
			eventStreamChan <- event
		}
	}

	go readEvent(oldES, oldEventStream, oldErrChan)
	go readEvent(instanceES, newEventStream, newErrChan)

	ret := &multierror.Error{}
	var event models.Event
	var lrpEvent LRPEvent
	for {
		var err error
		select {
		case e := <-oldEventStream:
			switch e.EventType() {
			case models.EventTypeActualLRPCreated, models.EventTypeActualLRPChanged, models.EventTypeActualLRPRemoved:
				event = e
			default:
				continue
			}
		case event = <-newEventStream:
		case err = <-oldErrChan:
			multierror.Append(ret, err)
		case err = <-newErrChan:
			multierror.Append(ret, err)
		}

		if len(ret.Errors) >= 2 {
			for _, err := range ret.Errors {
				if err != io.EOF {
					return ret.ErrorOrNil()
				}
			}
			return nil
		}

		if err != nil {
			continue
		}

		lrpEvent.Type = event.EventType()
		lrpEvent.Data = event
		err = encoder.Encode(lrpEvent)
		if err != nil {
			logger.Error("failed-to-marshal", err)
		}
	}
}

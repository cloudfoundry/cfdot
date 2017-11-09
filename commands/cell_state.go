package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/cfdot/commands/helpers"
	"code.cloudfoundry.org/rep"
	"github.com/spf13/cobra"
)

var cellStateCmd = &cobra.Command{
	Use:   "cell-state CELL_ID",
	Short: "Show the specified cell state",
	Long:  "Show the cell state specified by the given cell id",
	RunE:  cellState,
}

func init() {
	AddTLSFlags(cellStateCmd)
	RootCmd.AddCommand(cellStateCmd)
}

func cellState(cmd *cobra.Command, args []string) error {
	err := ValidateCellStateArguments(args)
	if err != nil {
		return NewCFDotValidationError(cmd, err)
	}

	bbsClient, err := helpers.NewBBSClient(cmd, clientConfig)
	if err != nil {
		return NewCFDotError(cmd, err)
	}
	cellRegistration, err := FetchCellRegistration(bbsClient, args[0])
	if err != nil {
		return NewCFDotError(cmd, err)
	}

	repClient, err := helpers.NewRepClient(cmd, cellRegistration.RepAddress, cellRegistration.RepUrl, clientConfig)
	if err != nil {
		return NewCFDotError(cmd, err)
	}

	err = FetchCellState(
		cmd.OutOrStdout(),
		cmd.OutOrStderr(),
		repClient,
	)
	if err != nil {
		return NewCFDotComponentError(cmd, fmt.Errorf("Rep error: %s", err.Error()))
	}

	return nil
}

func ValidateCellStateArguments(args []string) error {
	switch {
	case len(args) > 1:
		return errExtraArguments
	case len(args) < 1:
		return errMissingArguments
	default:
		return nil
	}
}

func FetchCellRegistration(bbsClient bbs.Client, cellId string) (*models.CellPresence, error) {
	logger := globalLogger.Session("fetch-cell-presence")

	cells, err := bbsClient.Cells(logger)
	if err != nil {
		return nil, err
	}

	for _, cell := range cells {
		if cell.CellId == cellId {
			return cell, nil
		}
	}

	return nil, errors.New("Cell not found")
}

func FetchCellState(stdout, stderr io.Writer, repClient rep.Client) error {
	logger := globalLogger.Session("cell-state")
	encoder := json.NewEncoder(stdout)

	state, err := repClient.State(logger)
	if err != nil {
		logger.Error("failed-to-fetch-cell-state", err)
		return err
	}

	err = encoder.Encode(state)
	if err != nil {
		logger.Error("failed-to-marshal", err)
		return err
	}
	return nil
}

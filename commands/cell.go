package commands

import (
	"encoding/json"
	"errors"
	"io"

	"code.cloudfoundry.org/bbs"

	"github.com/spf13/cobra"
)

var cellCmd = &cobra.Command{
	Use:   "cell CELL_ID",
	Short: "Show the specified cell presence",
	Long:  "Show the cell presence specified by the given cell id",
	RunE:  cell,
}

func init() {
	AddBBSFlags(cellCmd)
	RootCmd.AddCommand(cellCmd)
}

func cell(cmd *cobra.Command, args []string) error {
	err := ValidateCellArguments(args)
	if err != nil {
		return NewCFDotValidationError(cmd, err)
	}

	bbsClient, err := newBBSClient(cmd)
	if err != nil {
		return NewCFDotError(cmd, err)
	}

	err = Cell(
		cmd.OutOrStdout(),
		cmd.OutOrStderr(),
		bbsClient,
		args[0],
	)
	if err != nil {
		return NewCFDotError(cmd, err)
	}

	return nil
}

func ValidateCellArguments(args []string) error {
	switch {
	case len(args) > 1:
		return errExtraArguments
	case len(args) < 1:
		return errMissingArguments
	default:
		return nil
	}
}

func Cell(stdout, stderr io.Writer, bbsClient bbs.Client, cellId string) error {
	logger := globalLogger.Session("cell-presence")

	encoder := json.NewEncoder(stdout)

	cells, err := bbsClient.Cells(logger)
	if err != nil {
		return err
	}

	for _, cell := range cells {
		if cell.CellId == cellId {
			err = encoder.Encode(cell)
			if err != nil {
				logger.Error("failed-to-marshal", err)
			}

			return err
		}
	}

	return errors.New("Cell not found")
}

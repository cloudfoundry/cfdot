package commands

import (
	"encoding/json"
	"io"

	"code.cloudfoundry.org/bbs"

	"github.com/spf13/cobra"
)

var cellsCmd = &cobra.Command{
	Use:   "cells",
	Short: "List registered cell presences",
	Long:  "List registered cell presences from the BBS",
	RunE:  cells,
}

func init() {
	AddBBSFlags(cellsCmd)

	RootCmd.AddCommand(cellsCmd)
}

func cells(cmd *cobra.Command, args []string) error {
	err := ValidateCellsArguments(args)
	if err != nil {
		return NewCFDotValidationError(cmd, err)
	}

	bbsClient, err := newBBSClient(cmd)
	if err != nil {
		return NewCFDotError(cmd, err)
	}

	err = Cells(
		cmd.OutOrStdout(),
		cmd.OutOrStderr(),
		bbsClient,
	)
	if err != nil {
		return NewCFDotError(cmd, err)
	}

	return nil
}

func ValidateCellsArguments(args []string) error {
	if len(args) > 0 {
		return errExtraArguments
	}
	return nil
}

func Cells(stdout, stderr io.Writer, bbsClient bbs.Client) error {
	logger := globalLogger.Session("cell-presences")

	encoder := json.NewEncoder(stdout)

	cellPresences, err := bbsClient.Cells(logger)
	if err != nil {
		return err
	}

	for _, cellPresence := range cellPresences {
		err = encoder.Encode(cellPresence)
		if err != nil {
			logger.Error("failed-to-marshal", err)
		}
	}

	return nil
}

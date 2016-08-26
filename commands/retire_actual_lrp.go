package commands

import (
	"errors"
	"io"
	"strconv"

	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/bbs/models"
	"github.com/spf13/cobra"
)

var retireActualLRPCmd = &cobra.Command{
	Use:   "retire-actual-lrp <process-guid> <index>",
	Short: "Retire actual LRP by index and process guid",
	Long:  "Retire actual LRP by index and process guid",
	RunE:  retireActualLRP,
}
var errMissingArguments = errors.New("Missing arguments")
var errInvalidIndex = errors.New("Index should be a valid integer")
var errInvalidProcessGuid = errors.New("Process guid should be non empty string")

func init() {
	AddBBSFlags(retireActualLRPCmd)
	retireActualLRPCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		return BBSPrehook(cmd, args)
	}
	RootCmd.AddCommand(retireActualLRPCmd)
}

func retireActualLRP(cmd *cobra.Command, args []string) error {
	var err error
	var bbsClient bbs.Client

	if len(args) < 2 {
		return NewCFDotValidationError(cmd, errMissingArguments)
	}

	processGuid := args[0]
	if processGuid == "" {
		return NewCFDotValidationError(cmd, errInvalidProcessGuid)
	}

	indexInt64, err := strconv.ParseInt(args[1], 10, 32)
	if err != nil {
		return NewCFDotValidationError(cmd, errInvalidIndex)
	}
	index := int32(indexInt64)

	bbsClient, err = newBBSClient(cmd)
	if err != nil {
		return NewCFDotError(cmd, err)
	}

	err = RetireActualLRP(cmd.OutOrStdout(), cmd.OutOrStderr(), bbsClient, args, processGuid, index)
	if err != nil {
		return NewCFDotError(cmd, err)
	}

	return nil
}

func RetireActualLRP(stdout, stderr io.Writer, bbsClient bbs.Client, args []string, processGuid string, index int32) error {
	logger := globalLogger.Session("retireActualLRP")

	desiredLRP, err := bbsClient.DesiredLRPByProcessGuid(logger, processGuid)
	if err != nil {
		return err
	}

	actualLRPKey := models.ActualLRPKey{ProcessGuid: processGuid, Index: index, Domain: desiredLRP.Domain}
	err = bbsClient.RetireActualLRP(logger, &actualLRPKey)
	if err != nil {
		return err
	}

	return nil
}

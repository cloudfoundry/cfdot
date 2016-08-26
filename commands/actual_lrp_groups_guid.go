package commands

import (
	"encoding/json"
	"errors"
	"io"

	"code.cloudfoundry.org/bbs"

	"github.com/spf13/cobra"
)

var actualLRPGroupsByProcessGuidCmd = &cobra.Command{
	Use:   "actual-lrp-groups-for-guid <process-guid>",
	Short: "List actual LRP groups for a process guid",
	Long:  "List actual LRP groups from the BBS for a process guid",
	RunE:  actualLRPGroupsByProcessGuid,
}

var errMissingProcessGuid = errors.New("No process-guid given")

func init() {
	AddBBSFlags(actualLRPGroupsByProcessGuidCmd)
	actualLRPGroupsByProcessGuidCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		return BBSPrehook(cmd, args)
	}
	RootCmd.AddCommand(actualLRPGroupsByProcessGuidCmd)
}

func actualLRPGroupsByProcessGuid(cmd *cobra.Command, args []string) error {
	var err error
	var bbsClient bbs.Client

	bbsClient, err = newBBSClient(cmd)
	if err != nil {
		return NewCFDotError(cmd, err)
	}

	err = ActualLRPGroupsByProcessGuid(cmd.OutOrStdout(), cmd.OutOrStderr(), bbsClient, args)
	if err != nil {
		return NewCFDotError(cmd, err)
	}

	return nil
}

func ActualLRPGroupsByProcessGuid(stdout, stderr io.Writer, bbsClient bbs.Client, args []string) error {
	logger := globalLogger.Session("actualLRPGroupsByProcessGuid")

	if len(args) == 0 || args[0] == "" {
		return errMissingProcessGuid
	}
	processGuid := args[0]

	encoder := json.NewEncoder(stdout)
	actualLRPGroups, err := bbsClient.ActualLRPGroupsByProcessGuid(logger, processGuid)
	if err != nil {
		return err
	}

	for _, actualLRPGroup := range actualLRPGroups {
		encoder.Encode(actualLRPGroup)
	}
	return nil
}

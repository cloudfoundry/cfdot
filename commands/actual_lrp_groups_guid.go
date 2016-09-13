package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"code.cloudfoundry.org/bbs"

	"github.com/spf13/cobra"
)

// flags
var (
	actualLRPGroupsGuidIndexFlag = ""
)

// errors
var (
	errMissingProcessGuid = errors.New("No process-guid given")
)

var actualLRPGroupsByProcessGuidCmd = &cobra.Command{
	Use:   "actual-lrp-groups-for-guid <process-guid>",
	Short: "List actual LRP groups for a process guid",
	Long:  fmt.Sprintf("List actual LRP groups from the BBS for a given process guid. Process guids can be obtained by running %s actual-lrp-groups", os.Args[0]),
	RunE:  actualLRPGroupsByProcessGuid,
}

func init() {
	AddBBSFlags(actualLRPGroupsByProcessGuidCmd)
	actualLRPGroupsByProcessGuidCmd.Flags().StringVarP(&actualLRPGroupsGuidIndexFlag, "index", "i", "", "retrieve actual lrp for the given index")
	actualLRPGroupsByProcessGuidCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 || args[0] == "" {
			return NewCFDotValidationError(cmd, errMissingProcessGuid)
		}

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

	if actualLRPGroupsGuidIndexFlag == "" {
		err = ActualLRPGroupsByProcessGuid(cmd.OutOrStdout(), cmd.OutOrStderr(), bbsClient, args)
	} else {
		indexAsInt, err := ValidatePositiveIntegerForFlag("index", actualLRPGroupsGuidIndexFlag, cmd)
		if err != nil {
			return err
		}
		err = ActualLRPGroupsByProcessGuidAndIndex(cmd.OutOrStdout(), cmd.OutOrStderr(), bbsClient, args, indexAsInt)
	}
	if err != nil {
		return NewCFDotError(cmd, err)
	}

	return nil
}

func ActualLRPGroupsByProcessGuidAndIndex(stdout, stderr io.Writer, bbsClient bbs.Client, args []string, index int) error {
	logger := globalLogger.Session("actualLRPGroupsByProcessGuid")

	processGuid := args[0]
	actualLRPGroup, err := bbsClient.ActualLRPGroupByProcessGuidAndIndex(logger, processGuid, index)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(stdout)
	encoder.Encode(actualLRPGroup)
	return nil
}

func ActualLRPGroupsByProcessGuid(stdout, stderr io.Writer, bbsClient bbs.Client, args []string) error {
	logger := globalLogger.Session("actualLRPGroupsByProcessGuid")

	processGuid := args[0]

	actualLRPGroups, err := bbsClient.ActualLRPGroupsByProcessGuid(logger, processGuid)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(stdout)
	for _, actualLRPGroup := range actualLRPGroups {
		encoder.Encode(actualLRPGroup)
	}
	return nil
}

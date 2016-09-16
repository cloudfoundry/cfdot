package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

	"code.cloudfoundry.org/bbs"

	"github.com/spf13/cobra"
)

// flags
var (
	actualLRPGroupsGuidIndexFlag string
)

var actualLRPGroupsByProcessGuidCmd = &cobra.Command{
	Use:   "actual-lrp-groups-for-guid <process-guid>",
	Short: "List actual LRP groups for a process guid",
	Long:  fmt.Sprintf("List actual LRP groups from the BBS for a given process guid. Process guids can be obtained by running %s actual-lrp-groups", os.Args[0]),
	RunE:  actualLRPGroupsByProcessGuid,
}

func init() {
	AddBBSFlags(actualLRPGroupsByProcessGuidCmd)
	actualLRPGroupsByProcessGuidCmd.PreRunE = BBSPrehook

	// String flag because logic for optional int flag is not clear
	actualLRPGroupsByProcessGuidCmd.Flags().StringVarP(&actualLRPGroupsGuidIndexFlag, "index", "i", "", "retrieve actual lrp for the given index")
	RootCmd.AddCommand(actualLRPGroupsByProcessGuidCmd)
}

func actualLRPGroupsByProcessGuid(cmd *cobra.Command, args []string) error {
	processGuid, index, err := ValidateActualLRPGroupsForGuidArgs(args, actualLRPGroupsGuidIndexFlag)
	if err != nil {
		return NewCFDotValidationError(cmd, err)
	}

	bbsClient, err := newBBSClient(cmd)
	if err != nil {
		return NewCFDotError(cmd, err)
	}

	err = ActualLRPGroupsForGuid(cmd.OutOrStdout(), cmd.OutOrStderr(), bbsClient, processGuid, index)
	if err != nil {
		return NewCFDotError(cmd, err)
	}

	return nil
}

func ValidateActualLRPGroupsForGuidArgs(args []string, indexFlag string) (string, int, error) {
	if len(args) < 1 {
		return "", 0, errors.New("no process guid specified")
	}

	if len(args) > 1 {
		return "", 0, errors.New("too many arguments specified")
	}

	if args[0] == "" {
		return "", 0, errors.New("process guid cannot be an empty string")
	}

	index := -1
	if indexFlag != "" {
		var err error
		index, err = strconv.Atoi(indexFlag)
		if err != nil || index < 0 {
			return "", 0, errors.New("index must be an integer greater than 0")
		}
	}

	return args[0], index, nil
}

func ActualLRPGroupsForGuid(stdout, stderr io.Writer, bbsClient bbs.Client, processGuid string, index int) error {
	logger := globalLogger.Session("actual-lrp-groups-for-guid")

	encoder := json.NewEncoder(stdout)
	if index < 0 {
		actualLRPGroups, err := bbsClient.ActualLRPGroupsByProcessGuid(logger, processGuid)
		if err != nil {
			return err
		}

		return encoder.Encode(actualLRPGroups)
	} else {
		actualLRPGroup, err := bbsClient.ActualLRPGroupByProcessGuidAndIndex(logger, processGuid, index)
		if err != nil {
			return err
		}

		return encoder.Encode(actualLRPGroup)
	}
}

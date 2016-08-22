package commands

import (
	"encoding/json"
	"io"

	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/bbs/models"

	"github.com/spf13/cobra"
)

func init() {
	AddBBSFlags(actualLRPGroupsCmd)
	actualLRPGroupsCmd.PreRunE = BBSPrehook
	RootCmd.AddCommand(actualLRPGroupsCmd)
}

var actualLRPGroupsCmd = &cobra.Command{
	Use:   "actual-lrp-groups",
	Short: "List actual LRP groups",
	Long:  "List actual LRP groups from the BBS",
	RunE:  actualLRPGroups,
}

func actualLRPGroups(cmd *cobra.Command, args []string) error {
	var err error
	var bbsClient bbs.Client

	bbsClient, err = newBBSClient(cmd)
	if err != nil {
		return NewCFDotError(cmd, err)
	}

	err = ActualLRPGroups(cmd.OutOrStdout(), cmd.OutOrStderr(), bbsClient, args)
	if err != nil {
		return NewCFDotError(cmd, err)
	}

	return nil
}

func ActualLRPGroups(stdout, stderr io.Writer, bbsClient bbs.Client, args []string) error {
	logger := globalLogger.Session("actualLRPGroups")

	encoder := json.NewEncoder(stdout)
	actualLRPGroups, err := bbsClient.ActualLRPGroups(logger, models.ActualLRPFilter{})
	if err != nil {
		return err
	}

	for _, actualLRPGroup := range actualLRPGroups {
		encoder.Encode(actualLRPGroup)
	}

	return nil
}

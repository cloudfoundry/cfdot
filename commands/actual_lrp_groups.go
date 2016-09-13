package commands

import (
	"encoding/json"
	"io"

	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/bbs/models"

	"github.com/spf13/cobra"
)

// flags
var (
	actualLRPGroupsDomainFlag string
	actualLRPGroupsCellIdFlag string
)

var actualLRPGroupsCmd = &cobra.Command{
	Use:   "actual-lrp-groups",
	Short: "List actual LRP groups",
	Long:  "List actual LRP groups from the BBS",
	RunE:  actualLRPGroups,
}

func init() {
	AddBBSFlags(actualLRPGroupsCmd)
	actualLRPGroupsCmd.PreRunE = BBSPrehook
	actualLRPGroupsCmd.Flags().StringVarP(&actualLRPGroupsDomainFlag, "domain", "d", "", "retrieve only actual lrps for the given domain")
	actualLRPGroupsCmd.Flags().StringVarP(&actualLRPGroupsCellIdFlag, "cell-id", "c", "", "retrieve only actual lrps for the given cell id")
	RootCmd.AddCommand(actualLRPGroupsCmd)
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
	actualLRPFilter := models.ActualLRPFilter{}
	if actualLRPGroupsDomainFlag != "" {
		actualLRPFilter.Domain = actualLRPGroupsDomainFlag
	}

	if actualLRPGroupsCellIdFlag != "" {
		actualLRPFilter.CellID = actualLRPGroupsCellIdFlag
	}

	actualLRPGroups, err := bbsClient.ActualLRPGroups(logger, actualLRPFilter)
	if err != nil {
		return err
	}

	for _, actualLRPGroup := range actualLRPGroups {
		encoder.Encode(actualLRPGroup)
	}

	return nil
}

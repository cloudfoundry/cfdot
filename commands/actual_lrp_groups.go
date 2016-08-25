package commands

import (
	"encoding/json"
	"io"

	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/bbs/models"

	"github.com/spf13/cobra"
)

var (
	domain string
	cellId string
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
	actualLRPGroupsCmd.Flags().StringVarP(&domain, "domain", "d", "", "retrieve only actual lrps for the given domain")
	actualLRPGroupsCmd.Flags().StringVarP(&cellId, "cell-id", "c", "", "retrieve only actual lrps for the given cell id")
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
	if domain != "" {
		actualLRPFilter.Domain = domain
	}

	if cellId != "" {
		actualLRPFilter.CellID = cellId
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

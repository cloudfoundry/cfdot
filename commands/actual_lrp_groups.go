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
	bbsClient, err := newBBSClient(cmd)
	if err != nil {
		return NewCFDotError(cmd, err)
	}

	err = ActualLRPGroups(
		cmd.OutOrStdout(),
		cmd.OutOrStderr(),
		bbsClient,
		actualLRPGroupsDomainFlag,
		actualLRPGroupsCellIdFlag,
	)
	if err != nil {
		return NewCFDotError(cmd, err)
	}

	return nil
}

func ActualLRPGroups(stdout, stderr io.Writer, bbsClient bbs.Client, domain, cellID string) error {
	logger := globalLogger.Session("actual-lrp-groups")

	encoder := json.NewEncoder(stdout)

	actualLRPFilter := models.ActualLRPFilter{
		CellID: cellID,
		Domain: domain,
	}

	actualLRPGroups, err := bbsClient.ActualLRPGroups(logger, actualLRPFilter)
	if err != nil {
		return err
	}

	for _, actualLRPGroup := range actualLRPGroups {
		err = encoder.Encode(actualLRPGroup)
		if err != nil {
			logger.Error("failed-to-unmarshal", err)
		}
	}

	return nil
}

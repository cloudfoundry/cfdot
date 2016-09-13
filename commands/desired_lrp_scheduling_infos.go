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
	desiredLRPSchedulingInfoDomainFlag string
)

var desiredLRPSchedulingInfosCmd = &cobra.Command{
	Use:   "desired-lrp-scheduling-infos",
	Short: "List desired LRP scheduling infos",
	Long:  "List desired LRP scheduling infos from the BBS",
	RunE:  desiredLRPSchedulingInfos,
}

func init() {
	AddBBSFlags(desiredLRPSchedulingInfosCmd)
	desiredLRPSchedulingInfosCmd.PreRunE = BBSPrehook
	desiredLRPSchedulingInfosCmd.Flags().StringVarP(&desiredLRPSchedulingInfoDomainFlag, "domain", "d", "", "retrieve only scheduling infos for the given domain")
	RootCmd.AddCommand(desiredLRPSchedulingInfosCmd)
}

func desiredLRPSchedulingInfos(cmd *cobra.Command, args []string) error {
	var err error
	var bbsClient bbs.Client

	err = ValidateConflictingShortAndLongFlag("-d", "--domain", cmd)
	if err != nil {
		return err
	}

	bbsClient, err = newBBSClient(cmd)
	if err != nil {
		return NewCFDotError(cmd, err)
	}

	err = DesiredLRPSchedulingInfos(cmd.OutOrStdout(), cmd.OutOrStderr(), bbsClient, args)
	if err != nil {
		return NewCFDotError(cmd, err)
	}

	return nil
}

func DesiredLRPSchedulingInfos(stdout, stderr io.Writer, bbsClient bbs.Client, args []string) error {
	logger := globalLogger.Session("desiredLRPSchedulingInfos")

	encoder := json.NewEncoder(stdout)
	desiredLRPFilter := models.DesiredLRPFilter{}

	if desiredLRPSchedulingInfoDomainFlag != "" {
		desiredLRPFilter.Domain = desiredLRPSchedulingInfoDomainFlag
	}

	desiredLRPSchedulingInfos, err := bbsClient.DesiredLRPSchedulingInfos(logger, desiredLRPFilter)
	if err != nil {
		return err
	}

	for _, desiredLRPSchedulingInfo := range desiredLRPSchedulingInfos {
		encoder.Encode(desiredLRPSchedulingInfo)
	}

	return nil
}

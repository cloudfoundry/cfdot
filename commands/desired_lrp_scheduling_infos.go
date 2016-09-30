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
	desiredLRPSchedulingInfosDomainFlag string
)

var desiredLRPSchedulingInfosCmd = &cobra.Command{
	Use:   "desired-lrp-scheduling-infos",
	Short: "List desired LRP scheduling infos",
	Long:  "List desired LRP scheduling infos from the BBS",
	RunE:  desiredLRPSchedulingInfos,
}

func init() {
	AddBBSFlags(desiredLRPSchedulingInfosCmd)
	desiredLRPSchedulingInfosCmd.Flags().StringVarP(&desiredLRPSchedulingInfosDomainFlag, "domain", "d", "", "retrieve only scheduling infos for the given domain")
	RootCmd.AddCommand(desiredLRPSchedulingInfosCmd)
}

func desiredLRPSchedulingInfos(cmd *cobra.Command, args []string) error {
	err := ValidateConflictingShortAndLongFlag("-d", "--domain", cmd)
	if err != nil {
		return NewCFDotValidationError(cmd, err)
	}

	err = ValidateDesiredLRPSchedulingInfosArguments(args)
	if err != nil {
		return NewCFDotValidationError(cmd, err)
	}

	bbsClient, err := newBBSClient(cmd)
	if err != nil {
		return NewCFDotError(cmd, err)
	}

	err = DesiredLRPSchedulingInfos(cmd.OutOrStdout(), cmd.OutOrStderr(), bbsClient, desiredLRPSchedulingInfosDomainFlag)
	if err != nil {
		return NewCFDotError(cmd, err)
	}

	return nil
}

func ValidateDesiredLRPSchedulingInfosArguments(args []string) error {
	if len(args) > 0 {
		return errExtraArguments
	}
	return nil
}

func DesiredLRPSchedulingInfos(stdout, stderr io.Writer, bbsClient bbs.Client, domain string) error {
	logger := globalLogger.Session("desired-lrp-scheduling-infos")

	encoder := json.NewEncoder(stdout)
	desiredLRPFilter := models.DesiredLRPFilter{
		Domain: domain,
	}

	desiredLRPSchedulingInfos, err := bbsClient.DesiredLRPSchedulingInfos(logger, desiredLRPFilter)
	if err != nil {
		return err
	}

	for _, info := range desiredLRPSchedulingInfos {
		err = encoder.Encode(info)
		if err != nil {
			logger.Error("failed-to-marshal", err)
		}
	}

	return nil
}

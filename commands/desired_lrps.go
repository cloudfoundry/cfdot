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
	desiredLRPsDomainFlag string
)

var desiredLRPsCmd = &cobra.Command{
	Use:   "desired-lrps",
	Short: "List desired LRPs",
	Long:  "List desired LRPs from the BBS",
	RunE:  desiredLRPs,
}

func init() {
	AddBBSFlags(desiredLRPsCmd)
	desiredLRPsCmd.PreRunE = BBSPrehook
	desiredLRPsCmd.Flags().StringVarP(&desiredLRPsDomainFlag, "domain", "d", "", "retrieve only desired lrps for the given domain")
	RootCmd.AddCommand(desiredLRPsCmd)
}

func desiredLRPs(cmd *cobra.Command, args []string) error {
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

	err = DesiredLRPs(cmd.OutOrStdout(), cmd.OutOrStderr(), bbsClient, args)
	if err != nil {
		return NewCFDotError(cmd, err)
	}

	return nil
}

func DesiredLRPs(stdout, stderr io.Writer, bbsClient bbs.Client, args []string) error {
	logger := globalLogger.Session("desiredLRPs")

	encoder := json.NewEncoder(stdout)
	desiredLRPFilter := models.DesiredLRPFilter{}

	if desiredLRPsDomainFlag != "" {
		desiredLRPFilter.Domain = desiredLRPsDomainFlag
	}

	desiredLRPs, err := bbsClient.DesiredLRPs(logger, desiredLRPFilter)
	if err != nil {
		return err
	}

	for _, desiredLRP := range desiredLRPs {
		encoder.Encode(desiredLRP)
	}

	return nil
}

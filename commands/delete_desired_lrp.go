package commands

import (
	"io"

	"code.cloudfoundry.org/bbs"
	"github.com/spf13/cobra"
)

var deleteDesiredLRPCmd = &cobra.Command{
	Use:   "delete-desired-lrp PROCESS_GUID",
	Short: "Delete a desired LRP",
	Long:  "Delete a desired LRP with the given process guid.",
	RunE:  deleteDesiredLRP,
}

func init() {
	AddBBSFlags(deleteDesiredLRPCmd)
	RootCmd.AddCommand(deleteDesiredLRPCmd)
}

func deleteDesiredLRP(cmd *cobra.Command, args []string) error {
	processGuid, err := ValidateDeleteDesiredLRPArguments(args)
	if err != nil {
		return NewCFDotValidationError(cmd, err)
	}

	bbsClient, err := newBBSClient(cmd)
	if err != nil {
		return NewCFDotError(cmd, err)
	}

	err = DeleteDesiredLRP(cmd.OutOrStdout(), cmd.OutOrStderr(), bbsClient, processGuid)
	if err != nil {
		return NewCFDotError(cmd, err)
	}

	return nil
}

func ValidateDeleteDesiredLRPArguments(args []string) (string, error) {
	if len(args) == 0 {
		return "", errMissingArguments
	}

	if len(args) > 1 {
		return "", errExtraArguments
	}

	if args[0] == "" {
		return "", errInvalidProcessGuid
	}

	return args[0], nil
}

func DeleteDesiredLRP(stdout, stderr io.Writer, bbsClient bbs.Client, processGuid string) error {
	logger := globalLogger.Session("delete-desired-lrp")

	err := bbsClient.RemoveDesiredLRP(logger, processGuid)
	if err != nil {
		return err
	}

	return nil
}

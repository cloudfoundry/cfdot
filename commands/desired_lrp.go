package commands

import (
	"encoding/json"
	"errors"
	"io"

	"code.cloudfoundry.org/bbs"
	"github.com/spf13/cobra"
)

var desiredLRPCmd = &cobra.Command{
	Use:   "desired-lrp",
	Short: "Show the specified desired LRP",
	Long:  "Show the desired LRP specified by the given process GUID",
	RunE:  desiredLRP,
}

func init() {
	AddBBSFlags(desiredLRPCmd)
	desiredLRPCmd.PreRunE = BBSPrehook
	RootCmd.AddCommand(desiredLRPCmd)
}

func desiredLRP(cmd *cobra.Command, args []string) error {
	processGuid, err := ValidateDesiredLRPArguments(args)
	if err != nil {
		return NewCFDotValidationError(cmd, err)
	}

	bbsClient, err := newBBSClient(cmd)
	if err != nil {
		return NewCFDotError(cmd, err)
	}

	err = DesiredLRP(cmd.OutOrStdout(), cmd.OutOrStderr(), bbsClient, processGuid)
	if err != nil {
		return NewCFDotError(cmd, err)
	}

	return nil
}

func ValidateDesiredLRPArguments(args []string) (string, error) {
	if len(args) == 0 {
		return "", errors.New("no process guid specified")
	}

	if len(args) > 1 {
		return "", errors.New("too many arguments specified")
	}

	if (args[0]) == "" {
		return "", errors.New("process guid cannot be an empty string")
	}

	return args[0], nil
}

func DesiredLRP(stdout, stderr io.Writer, bbsClient bbs.Client, processGuid string) error {
	logger := globalLogger.Session("desiredLRP")

	desiredLRP, err := bbsClient.DesiredLRPByProcessGuid(logger, processGuid)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(stdout)
	return encoder.Encode(desiredLRP)
}

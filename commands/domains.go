package commands

import (
	"encoding/json"
	"io"

	"code.cloudfoundry.org/bbs"

	"github.com/spf13/cobra"
)

var domainsCmd = &cobra.Command{
	Use:   "domains",
	Short: "List domains",
	Long:  "List fresh domains from the BBS",
	RunE:  domains,
}

func init() {
	AddBBSFlags(domainsCmd)
	RootCmd.AddCommand(domainsCmd)
}

func domains(cmd *cobra.Command, args []string) error {
	err := ValidateDomainsArguments(args)
	if err != nil {
		return NewCFDotValidationError(cmd, err)
	}

	bbsClient, err := newBBSClient(cmd)
	if err != nil {
		return NewCFDotError(cmd, err)
	}

	err = Domains(cmd.OutOrStdout(), cmd.OutOrStderr(), bbsClient)
	if err != nil {
		return NewCFDotError(cmd, err)
	}

	return nil
}

func ValidateDomainsArguments(args []string) error {
	if len(args) > 0 {
		return errExtraArguments
	}
	return nil
}

func Domains(stdout, stderr io.Writer, bbsClient bbs.Client) error {
	logger := globalLogger.Session("domains")

	encoder := json.NewEncoder(stdout)
	domains, err := bbsClient.Domains(logger)
	if err != nil {
		return err
	}

	for _, domain := range domains {
		err = encoder.Encode(domain)
		if err != nil {
			logger.Error("failed-to-marshal", err)
		}
	}

	return nil
}

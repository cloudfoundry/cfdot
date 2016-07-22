package commands

import (
	"encoding/json"
	"errors"
	"io"

	"code.cloudfoundry.org/bbs"

	"github.com/spf13/cobra"
)

func init() {
	addBBSFlags(domainsCmd)
	RootCmd.AddCommand(domainsCmd)
}

var domainsCmd = &cobra.Command{
	Use:   "domains",
	Short: "List domains",
	Long:  "List the available domains",
	Run:   domains,
}

func domains(cmd *cobra.Command, args []string) {
	if bbsURL == "" {
		reportErr(cmd.OutOrStderr(), errors.New("the required flag '--bbsURL' was not specified"))
	}

	bbsClient := bbs.NewClient(bbsURL)

	err := Domains(cmd.OutOrStdout(), cmd.OutOrStderr(), bbsClient, args)
	if err != nil {
		reportErr(cmd.OutOrStderr(), err)
	}
}

func Domains(stdout, stderr io.Writer, bbsClient bbs.Client, args []string) error {
	logger = logger.Session("domains")

	encoder := json.NewEncoder(stdout)
	domains, err := bbsClient.Domains(logger)
	if err != nil {
		return err
	}

	for _, domain := range domains {
		encoder.Encode(domain)
	}

	return nil
}
